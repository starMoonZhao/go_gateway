package public

const (
	ValidatorKey        = "ValidatorKey"
	TranslatorKey       = "TranslatorKey"
	AdminSessionInfoKey = "AdminSessionInfoKey"

	//服务类型
	LoadTypeHTTP = 0
	LoadTypeTCP  = 1
	LoadTypeGRPC = 2

	//http服务的规则类型
	HTTPRuleTypePrefixURL = 0
	HTTPRuleTypeDomain    = 1

	//http服务是否使用https
	HTTPDontNeedHttps = 0
	HTTPNeedHttps     = 1

	//流量统计数据在redis中存储的前缀标识
	RedisFlowDayKey  = "flow_day_count"
	RedisFlowHourKey = "flow_hour_count"

	//流量统计器ID前缀
	FlowTotal   = "flow_total"   //全站流量
	FlowService = "flow_service" //服务流量
	FlowApp     = "flow_app"     //租户流量
)

var (
	//定义服务类型字典
	LoadTypeMap = map[int]string{
		LoadTypeHTTP: "HTTP",
		LoadTypeTCP:  "TCP",
		LoadTypeGRPC: "GRPC",
	}
)
