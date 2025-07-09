package ulid

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/oklog/ulid/v2"
)

type Prefix string

const (
	UserPrefix      Prefix = "usr"
	SessionPrefix   Prefix = "ses"
	SafePrefix      Prefix = "saf"
	RolePrefix      Prefix = "rol"
	OperationPrefix Prefix = "op"
	ConfigPrefix    Prefix = "cfg"
	CAPrefix        Prefix = "ca"
	CyberArkInstancePrefix Prefix = "cai"
)

func New(prefix Prefix) string {
	t := time.Now()
	entropy := ulid.Monotonic(rand.Reader, 0)
	id := ulid.MustNew(ulid.Timestamp(t), entropy)
	return fmt.Sprintf("%s_%s", prefix, id.String())
}

func IsValid(id string, prefix Prefix) bool {
	expected := fmt.Sprintf("%s_", prefix)
	if len(id) != len(expected)+26 {
		return false
	}
	if id[:len(expected)] != expected {
		return false
	}
	_, err := ulid.Parse(id[len(expected):])
	return err == nil
}