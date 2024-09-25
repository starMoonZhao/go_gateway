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
)
