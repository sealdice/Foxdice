package satori

type AssetsManager interface {
	Get(url string) string
	Add(url string) string
}

type assetsManager struct{}

func (a *assetsManager) Get(url string) string {
	// TODO implement me
	panic("implement me")
}

func (a *assetsManager) Add(url string) string {
	// TODO implement me
	panic("implement me")
}

type assets struct {
	id     string
	origin string
	url    string
}
