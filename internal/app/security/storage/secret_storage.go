package storage

type ISecretStorage interface {
	StorePassword(password string) error
	RetrievePassword() (string, error)
}
