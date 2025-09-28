package interfaces

type InitArgs struct {
	ConfigFile     string
	IsConfidential bool
	Port           string
	Bind           string
	Address        string
	PubCertKeyPath string
	PubKeyPath     string
	Pwd            string
}
