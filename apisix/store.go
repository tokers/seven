package apisix

import (
	"github.com/gxthrj/seven/DB"
	"github.com/gxthrj/apisix-types/pkg/apis/apisix/v1"
	"fmt"
	"github.com/gxthrj/seven/conf"
	"github.com/golang/glog"
)



// FindRoute find current route in memDB
func FindRoute(route *v1.Route) (*v1.Route,error){
	db := &DB.RouteRequest{Name: *(route.Name)}
	currentRoute, _ := db.FindByName()
	if currentRoute != nil {
		return currentRoute, nil
	} else {
		// find from apisix
		if routes, err := ListRoute(); err != nil {
			// todo log error
		} else {
			for _, r := range routes {
				if r.Name !=nil && *r.Name == *route.Name {
					// insert to memDB
					db := &DB.RouteDB{Routes: []*v1.Route{r}}
					db.Insert()
					// return
					return r, nil
				}
			}
		}

	}
	return nil, fmt.Errorf("NOT FOUND")
}
// FindUpstreamByName find upstream from memDB,
// if Not Found, find upstream from apisix
func FindUpstreamByName(name string) (*v1.Upstream, error){
	ur := &DB.UpstreamRequest{Name: name}
	currentUpstream, _ := ur.FindByName()
	if currentUpstream != nil {
		return currentUpstream, nil
	} else {
		// find upstream from apisix
		if upstreams, err := ListUpstream(); err != nil {
			// todo log error
		}else {
			for _, upstream := range upstreams {
				if upstream.Name != nil && *(upstream.Name) == name {
					// and save to memDB
					upstreamDB := &DB.UpstreamDB{Upstreams: []*v1.Upstream{upstream}}
					upstreamDB.InsertUpstreams()
					//InsertUpstreams([]*v1.Upstream{upstream})
					// return
					return upstream, nil
				}
			}
		}

	}
	return nil, nil
}

// FindServiceByName find service from memDB,
// if Not Found, find service from apisix
func FindServiceByName(name string) (*v1.Service, error){
	db := DB.ServiceRequest{Name: name}
	currentService, _ := db.FindByName()
	if currentService != nil {
		return currentService, nil
	}else {
		// find service from apisix
		if services, err := ListService(conf.BaseUrl); err != nil {
			// todo log error
			glog.Info(err.Error())
		}else {
			for _, s := range services {
				if s.Name != nil && *(s.Name) == name {
					// and save to memDB
					db := &DB.ServiceDB{Services: []*v1.Service{s}}
					db.Insert()
					// return
					return s, nil
				}
			}
		}
	}
	return nil, nil
}