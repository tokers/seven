package apisix

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"github.com/gxthrj/apisix-types/pkg/apis/apisix/v1"
	"github.com/gxthrj/seven/conf"
	"github.com/gxthrj/seven/utils"
	"strconv"
	"strings"
	"github.com/gxthrj/seven/DB"
)

// FindCurrentUpstream find upstream from memDB,
// if Not Found, find upstream from apisix
func FindCurrentUpstream(group, name string) (*v1.Upstream, error){
	ur := &DB.UpstreamRequest{Group: group, Name: name}
	currentUpstream, _ := ur.FindByName()
	if currentUpstream != nil {
		return currentUpstream, nil
	} else {
		// find upstream from apisix
		if upstreams, err := ListUpstream(group); err != nil {
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

// ListUpstream list upstream from etcd , convert to v1.Upstream
func ListUpstream(group string) ([]*v1.Upstream, error) {
	baseUrl := conf.FindUrl(group)
	url := baseUrl + "/upstreams"
	ret, _ := Get(url)
	var upstreamsResponse UpstreamsResponse
	if err := json.Unmarshal(ret, &upstreamsResponse); err != nil {
		return nil, fmt.Errorf("json转换失败")
	} else {
		upstreams := make([]*v1.Upstream, 0)
		for _, u := range upstreamsResponse.Upstreams.Upstreams {
			if n, err := u.convert(group); err == nil {
				upstreams = append(upstreams, n)
			} else {
				return nil, fmt.Errorf("upstream: %s 转换失败, %s", *u.UpstreamNodes.Desc, err.Error())
			}
		}
		return upstreams, nil
	}
}

//func IsExist(name string) (bool, error) {
//	if upstreams, err := ListUpstream(); err != nil {
//		return false, err
//	} else {
//		for _, upstream := range upstreams {
//			if *upstream.Name == name {
//				return true, nil
//			}
//		}
//		return false, nil
//	}
//}

func AddUpstream(upstream *v1.Upstream) (*UpstreamResponse, error) {
	baseUrl := conf.FindUrl(*upstream.Group)
	url := fmt.Sprintf("%s/upstreams", baseUrl)
	glog.Info(url)
	ur := convert2UpstreamRequest(upstream)
	if b, err := json.Marshal(ur); err != nil {
		return nil, err
	} else {
		if res, err := utils.Post(url, b); err != nil {
			return nil, err
		} else {
			var uRes UpstreamResponse
			if err = json.Unmarshal(res, &uRes); err != nil {
				glog.Errorf("json Unmarshal error: %s", err.Error())
				return nil, err
			} else {
				glog.Info(uRes)
				if uRes.Upstream.Key != nil {
					return &uRes, nil
				} else {
					return nil, fmt.Errorf("apisix upstream not expected response")
				}
			}
		}
	}
}

func UpdateUpstream(upstream *v1.Upstream) error {
	baseUrl := conf.FindUrl(*upstream.Group)
	url := fmt.Sprintf("%s/upstreams/%s", baseUrl, *upstream.ID)
	ur := convert2UpstreamRequest(upstream)
	if b, err := json.Marshal(ur); err != nil {
		return err
	} else {
		if _, err := utils.Patch(url, b); err != nil {
			return err
		} else {
			return nil
		}
	}
}

func convert2UpstreamRequest(upstream *v1.Upstream) *UpstreamRequest {
	nodes := make(map[string]int64)
	for _, u := range upstream.Nodes {
		nodes[*u.IP+":"+strconv.Itoa(*u.Port)] = int64(*u.Weight)
	}
	return &UpstreamRequest{
		LBType: *upstream.Type,
		HashOn: upstream.HashOn,
		Key:    upstream.Key,
		Desc:   *upstream.Name,
		Nodes:  nodes,
	}
}

// convert convert Upstream from etcd to v1.Upstream
func (u *Upstream) convert(group string) (*v1.Upstream, error) {
	// id
	keys := strings.Split(*u.Key, "/")
	id := keys[len(keys)-1]
	// Name
	name := u.UpstreamNodes.Desc
	// type
	LBType := u.UpstreamNodes.LBType
	// key
	key := u.Key
	// nodes
	nodes := make([]*v1.Node, 0)
	for k, v := range u.UpstreamNodes.Nodes {
		ks := strings.Split(k, ":")
		ip := ks[0]
		port, _ := strconv.Atoi(ks[1])
		weight := int(v)
		node := &v1.Node{IP: &ip, Port: &port, Weight: &weight}
		nodes = append(nodes, node)
	}

	return &v1.Upstream{ID: &id, Group: &group, Name: name, Type: LBType, Key: key, Nodes: nodes}, nil
}

type UpstreamsResponse struct {
	Upstreams Upstreams `json:"node"`
}

type UpstreamResponse struct {
	Action   string   `json:"action"`
	Upstream Upstream `json:"node"`
}

type Upstreams struct {
	Key       string     `json:"key"` // 用来定位upstreams 列表
	Upstreams []Upstream `json:"nodes"`
}

type Upstream struct {
	Key           *string       `json:"key"` // upstream key
	UpstreamNodes UpstreamNodes `json:"value"`
}

type UpstreamNodes struct {
	Nodes  map[string]int64 `json:"nodes"`
	Desc   *string          `json:"desc"` // upstream name  = k8s svc
	LBType *string          `json:"type"` // 负载均衡类型
}

//{"type":"roundrobin","nodes":{"10.244.10.11:8080":100},"desc":"somesvc"}
type UpstreamRequest struct {
	LBType string           `json:"type"`
	HashOn *string          `json:"hash_on,omitempty"`
	Key    *string          `json:"key,omitempty"`
	Nodes  map[string]int64 `json:"nodes"`
	Desc   string           `json:"desc"`
}
