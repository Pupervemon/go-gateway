package config

// RouteConfig 路由配置结构体
type RouteConfig struct {
	ID         string      `json:"id"`
	Order      int         `json:"order,omitempty"`
	URI        string      `json:"uri"`
	Predicates []Predicate `json:"predicates"`
	Filters    []Filter    `json:"filters"`
}

// Predicate 路由断言
type Predicate struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
}

// Filter 路由过滤器
type Filter struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
}

// RoutesConfig 路由配置集合
type RoutesConfig struct {
	Routes []RouteConfig `json:"routes"`
}
