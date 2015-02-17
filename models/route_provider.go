package models

//go:generate counterfeiter -o fakes/fake_route_provider.go . RouteProvider
type RouteProvider interface {
	TiegoRoute(name string) string
	SSHRoute(name string) string
}

type routeProvider struct {
	appsDomain string
}

func NewRouteProvider(appsDomain string) *routeProvider {
	return &routeProvider{
		appsDomain: appsDomain,
	}
}

func (rp *routeProvider) TiegoRoute(name string) string {
	return "tiego-" + name + "." + rp.appsDomain
}

func (rp *routeProvider) SSHRoute(name string) string {
	return "ssh-" + name + "." + rp.appsDomain
}
