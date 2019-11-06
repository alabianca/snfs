package discovery

func MdnsStrategy(o ...Option) *MdnsService {
	return service(o...)
}
