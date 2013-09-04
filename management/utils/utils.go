package utils
import (
	"crypto/md5"
	"crypto/sha1"
    "fmt"
    "time"
)

func GetMd5Digest(seed string) string {
	h := md5.New()
	h.Write([]byte(seed))
	return fmt.Sprintf("%x", h.Sum([]byte("")))
}

func GetSha1Digest(seed string) string {
	h := sha1.New()
	h.Write([]byte(seed))
	return fmt.Sprintf("%x", h.Sum([]byte("")))
}

func GenerateSalt(seed string) string {
	return GetMd5Digest(time.Now().String() + seed)
}

func EncryptPassword(origin, salt string) string {
	return GetSha1Digest(salt + GetMd5Digest(origin) + salt)
}

func GenerateRegisterUrl(token string) string {
    return fmt.Sprintf("https://10.0.1.103:8080/register?token=%s", token)
}

func GenerateRegisterToken(emailAddr string) string {
    return GetMd5Digest(time.Now().String() + emailAddr + "MangoInviteString")
}

