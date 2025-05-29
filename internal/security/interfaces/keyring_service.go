package interfaces

type IKeyringService interface {
	StorePassword(password string) error
	RetrievePassword() (string, error)
}
