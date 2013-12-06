package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
    "encoding/hex"
    "net/url"
	"fmt"
    "strconv"
	"time"
	"Mango/management/models"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"labix.org/v2/mgo"
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
    if beego.HttpTLS { 
	    return fmt.Sprintf("https://%s/register?token=%s", beego.AppConfig.String("apphost"), token)
    }
	return fmt.Sprintf("http://%s/register?token=%s", beego.AppConfig.String("apphost"), token)
}

func GenerateRegisterToken(emailAddr string) string {
	return GetMd5Digest(time.Now().String() + emailAddr + "MangoInviteString")
}

func EncryptStringInAES(origin, key string) string {
	c, err := aes.NewCipher([]byte(key))
	if err != nil {
		panic(err)
	}
	var iv = []byte(base64.StdEncoding.EncodeToString([]byte(key)))[:aes.BlockSize]
	cfb := cipher.NewCFBEncrypter(c, iv)
	cipherText := make([]byte, len([]byte(origin)))
	cfb.XORKeyStream(cipherText, []byte(origin))
	return fmt.Sprintf("%x", cipherText)
}

func DecryptStringInAES(cipherText, key string) string {
	c, err := aes.NewCipher([]byte(key))
	if err != nil {
		panic(err)
	}
    cipherBytes, _  := hex.DecodeString(cipherText)
	var iv = []byte(base64.StdEncoding.EncodeToString([]byte(key)))[:aes.BlockSize]
	cfbdec := cipher.NewCFBDecrypter(c, iv)
	decrypted := make([]byte, len(cipherBytes))
	cfbdec.XORKeyStream(decrypted, cipherBytes)
	return fmt.Sprintf("%s", decrypted)
}

func GetTheKey() string {
	o := orm.NewOrm()
	key := models.MPKey{Id: 1}
	o.Read(&key)
	return GetMd5Digest("zningsaidwnf23ic3" + key.DataKey + "X@#$#@!324f2")
}

func NewRegisterMail(emailAddr, token string) *models.MangoMail {
	to := make([]string, 0)
	to = append(to, emailAddr)
	url := GenerateRegisterUrl(token)
	m := &models.MangoMail{
		FromVar:    "Guoku <noreply@post.guoku.com>",
		ToVar:      to,
		SubjectVar: "Mango Registration URL",
		HtmlVar:    fmt.Sprintf("<a href='%s'>点此进入</a>", url),
	}
	return m
}

func GetUploadItemParams(item *models.TaobaoItemStd, params *url.Values, matchedGuokuCid int) {
        params.Add("taobao_id", strconv.Itoa(item.NumIid))
        params.Add("cid", strconv.Itoa(item.Cid))
        params.Add("taobao_title", item.Title)
        params.Add("taobao_shop_nick", item.Nick)
        params.Add("taobao_price", fmt.Sprintf("%f", item.Price))
        itemImgs := item.ItemImgs
        if itemImgs != nil && len(itemImgs) > 0 {
            params.Add("chief_image_url", itemImgs[0])
            for i, _ := range itemImgs {
                params.Add("image_url", itemImgs[i])
            }
        }
        params.Add("category_id", strconv.Itoa(matchedGuokuCid))
}

func GetNewMongoSession() *mgo.Session {
	mongoHost := beego.AppConfig.String("mongohost")
    session, err := mgo.Dial(mongoHost)
    if err != nil {
        return nil
    }
    return session
}
