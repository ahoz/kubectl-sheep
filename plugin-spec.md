# kubectl-sheep — Plugin Spec

## 0. Naming

- **Plugin name (krew):** `sheep`
- **Binary:** `kubectl-sheep`
- **Invocation:** `kubectl sheep <command>`
- **Krew manifest short description** (for `kubectl krew search`):
  > "Fetch and manage kubeconfigs from Rancher-managed clusters"
- **Krew manifest long description** (discoverability via keywords like *rancher*, *kubeconfig*, *cluster*, *token*):
  > A kubectl plugin to manage multiple Rancher instances, list their downstream clusters, and fetch/refresh kubeconfigs individually or in bulk. Rancher API tokens can be stored either as plaintext or encrypted (passphrase-protected file backend), selectable per instance.

Name collision checked against the official `kubernetes-sigs/krew-index`: `sheep` is free (as of today — re-check before final submission).

Subcommands are deliberately kept purely technical (`instance`, `cluster`, `fetch-all`, ...) — no farm vocabulary in the command language, as that would hurt discoverability and readability of `--help`. The thematic reference lives exclusively in the plugin name itself.

---

## 1. Goal

A krew plugin (`kubectl sheep`) that:

- can manage multiple Rancher instances (name, URL, token)
- stores Rancher tokens **either** unencrypted (plain file, `0600`) or encrypted (`keyring.FileBackend`, passphrase-protected) — selectable per instance by the user, switchable at any time
- can list the clusters of an instance
- can fetch kubeconfigs for all clusters of an instance **or** for individually selected clusters
- can update (rotate) existing, locally stored kubeconfigs on demand
- offers a simple update flow if a Rancher token becomes invalid/expired

No D-Bus/Secret Service backend as a hard requirement (WSL-friendly), but optionally supportable later.

---

## 2. Data Model

### 2.1 Instance Config (non-secret)

`~/.config/kubectl-sheep/instances.yaml`

```yaml
instances:
  - name: prod
    url: https://rancher.prod.example.com
    insecureSkipVerify: false
    storage: encrypted     # plaintext | encrypted
  - name: dev
    url: https://rancher.dev.example.com
    insecureSkipVerify: true
    storage: plaintext
```

### 2.2 Credential Storage (secret, abstracted)

Interface implemented by both backends:

```go
type CredentialStore interface {
    Get(instance string) (token string, err error)
    Set(instance string, token string) error
    Delete(instance string) error
}
```

**PlaintextStore**
- File: `~/.config/kubectl-sheep/credentials.plain.yaml`, permissions `0600`
- simple `map[string]string` (instance -> token)

**EncryptedStore**
- `99designs/keyring`, `AllowedBackends: []keyring.BackendType{keyring.FileBackend}` (deliberately no SecretService as a mandatory fallback, see prior context re: WSL)
- Directory: `~/.config/kubectl-sheep/keys`
- Passphrase prompt via `keyring.TerminalPrompt` or a custom prompt (Bubbletea/Survey), 1 passphrase per keyring instance (not per Rancher token)
- Item key name: `rancher-token:<instance-name>`

**Switching storage mode** (`instance set-storage <name> --to=plaintext|encrypted`):
1. Read token from current backend
2. Write to new backend
3. Only after success: delete from old backend
4. Update `instances.yaml`

Important: the storage choice is **per instance**, not global — this cleanly covers "the user should be able to switch back and forth," including mixed setups (one instance plain, another encrypted).

### 2.3 Local Kubeconfig Storage

`~/.kube/sheep/<instance>/<cluster-id>.yaml` — one file per cluster, no mandatory automatic merge. Optional: `--merge` flag that mixes the contexts into `~/.kube/config` via `client-go/tools/clientcmd`, using the naming scheme `<instance>-<cluster-name>` to avoid collisions.

Additionally: a small metadata file per cluster (`<cluster-id>.meta.yaml`) with `fetchedAt`, `clusterId`, `clusterName` — the foundation for a later "how old is my local kubeconfig" display.

---

## 3. Rancher API Interaction

Only two endpoints are needed for the core functionality:

- `GET  /v3/clusters` — cluster list (name, ID, state) with `Authorization: Bearer <token>`
- `POST /v3/clusters/<id>?action=generateKubeconfig` — returns `.config` (the actual kubeconfig YAML as a string)

Token validation: a `GET /v3/clusters?limit=1` health check is sufficient — on `401`, the token is considered invalid → trigger the update flow.

---

## 4. CLI Commands (target state)

```
kubectl sheep instance add <name> --url=... [--storage=plaintext|encrypted] [--insecure]
kubectl sheep instance list
kubectl sheep instance remove <name>
kubectl sheep instance set-storage <name> --to=plaintext|encrypted
kubectl sheep instance update-token <name>          # invalid token -> set a new one

kubectl sheep cluster list <instance>
kubectl sheep cluster get <instance> <cluster>       # fetch a single cluster
kubectl sheep cluster refresh <instance> <cluster>   # targeted rotation

kubectl sheep fetch-all <instance> [--merge]
kubectl sheep refresh-all <instance> [--merge]       # re-fetch everything already stored locally
```

During `add`, the token is prompted for interactively (hidden input), not passed as a plaintext flag — otherwise it ends up in shell history.

---

## 5. Implementation Order (for the agent)

**Phase 0 — Scaffolding**
Cobra CLI skeleton, binary name correct per krew convention (`kubectl-sheep`), empty subcommands, `--help` texts. No logic yet.

**Phase 1 — Instance Config (no secrets)**
`instance add/list/remove` against `instances.yaml`. Validation: unique name, parseable URL. No token handling, no API calls — pure CRUD on YAML.

**Phase 2 — Credential Storage Abstraction**
`CredentialStore` interface + `PlaintextStore` + `EncryptedStore`. Unit tests for both, independent of the rest. `instance add` now prompts for the token and stores it per `--storage`. Implement `instance set-storage` (migration).

**Phase 3 — Rancher API Client**
Thin client for the two endpoints + token health check. Clean error handling for 401 (→ dedicated error type `ErrTokenInvalid`, caught later at the CLI layer).

**Phase 4 — Cluster List & Single Fetch**
`cluster list`, `cluster get` — fetch the kubeconfig for one cluster and store it under `~/.kube/sheep/<instance>/<cluster-id>.yaml`, including the metadata file.

**Phase 5 — Bulk Fetch**
`fetch-all` — iterate over all clusters of an instance, parallelizable (small worker pool, e.g. 5 concurrent), collect errors per cluster instead of aborting on the first failure.

**Phase 6 — Rotation/Refresh**
`cluster refresh` and `refresh-all` — essentially the same call as fetch, but explicitly operating over already-existing local files, with a diff hint ("kubeconfig for X updated, token expiry: ...", if extractable from the generated kubeconfig).

**Phase 7 — Token Invalidation**
Central error handling: any command that receives `ErrTokenInvalid` aborts with a clear message and suggests `instance update-token <name>`. `update-token` itself: prompt for the token, health-check it against the API, only store it after success (never blindly overwrite).

**Phase 8 — Merge into ~/.kube/config (optional)**
`--merge` flag for `fetch-all`/`cluster get`/`refresh-all`, using `clientcmd` to cleanly mix in the contexts, including a name-collision strategy.

**Phase 9 — Polish & Distribution**
Unify error messages, `krew` manifest (`plugin.yaml`, name `sheep`, see Section 0 for description/long description), README, ensure `insecureSkipVerify` is handled consistently (global vs. per-instance), tests for the API client's error paths (401, network errors, invalid cluster ID).

---

## 6. Deliberate Decisions / Non-Goals (v1)

- No automatic cron-like refresh — on-demand only, via explicit commands.
- No multi-user sharing of credential files — strictly local, per OS user.
- No SecretService/D-Bus requirement — `FileBackend` is the only encrypted path in v1, since it's WSL-friendly; genuine OS keychain support can be added later as a third option (`--storage=os-keychain`) without changing the interface.
- TTL display/expiry warning for generated kubeconfigs is "nice to have," not a core feature in v1 (Phase 6 lays the metadata-file groundwork for it, though).
- Plugin name (`sheep`) and command language (technical: `instance`/`cluster`/...) are deliberately kept separate — discoverability in the krew index relies on the manifest description, not on farm vocabulary in the subcommands.