package interfaces

type ISecretStorage interface {
	StorePassword(password string) error
	RetrievePassword() (string, error)
}
