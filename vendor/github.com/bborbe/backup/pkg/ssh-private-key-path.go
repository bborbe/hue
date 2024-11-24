package pkg

type SSHPrivateKey string

func (f SSHPrivateKey) String() string {
	return string(f)
}
