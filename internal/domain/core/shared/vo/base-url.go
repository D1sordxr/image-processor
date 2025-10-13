package vo

type BaseURL string

func (base BaseURL) String() string {
	return string(base)
}

func NewBaseURL(host, port string) BaseURL {
	return BaseURL("https://" + host + ":" + port)
}
