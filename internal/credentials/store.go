package credentials

// Store abstracts secret token storage per Rancher instance.
type Store interface {
	Get(instance string) (token string, err error)
	Set(instance string, token string) error
	Delete(instance string) error
}
