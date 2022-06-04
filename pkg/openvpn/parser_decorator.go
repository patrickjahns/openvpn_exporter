package openvpn

type ParserDecorator interface {
	DecorateParseFile(f func(statusfile string) (*Status, error)) func(statusfile string) (*Status, error)
}
