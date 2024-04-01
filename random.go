package orz

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Sign 排序+拼接+摘要
func Sign(a ...string) string {
	sort.Strings(a)
	data := []byte(strings.Join(a, ""))
	has := md5.Sum(data)
	return fmt.Sprintf("%x", has)
}

type passwordEncoder struct {
	cost int
}

func (b *passwordEncoder) Encode(password []byte) ([]byte, error) {
	return bcrypt.GenerateFromPassword(password, b.cost)
}

func (b *passwordEncoder) Match(hashedPassword, password []byte) error {
	return bcrypt.CompareHashAndPassword(hashedPassword, password)
}

var _passwordEncoder = passwordEncoder{
	cost: bcrypt.DefaultCost,
}

func PasswordEncode(password []byte) ([]byte, error) {
	return _passwordEncoder.Encode(password)
}

func PasswordMatch(hashedPassword, password []byte) error {
	return _passwordEncoder.Match(hashedPassword, password)
}

func GenPassword() string {
	rand.Seed(time.Now().UnixNano())
	digits := "0123456789"
	specials := "~=+%^*/()[]{}/!@#$?|"
	all := "ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		digits + specials
	length := 8
	buf := make([]byte, length)
	buf[0] = digits[rand.Intn(len(digits))]
	buf[1] = specials[rand.Intn(len(specials))]
	for i := 2; i < length; i++ {
		buf[i] = all[rand.Intn(len(all))]
	}
	rand.Shuffle(len(buf), func(i, j int) {
		buf[i], buf[j] = buf[j], buf[i]
	})
	return string(buf)
}

const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
const onlyNumber = "0123456789"

func GenRandomStr(n int) (randStr string) {
	charsLen := len(chars)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < n; i++ {
		randIndex := rand.Intn(charsLen)
		randStr += chars[randIndex : randIndex+1]
	}
	return randStr
}

func GenRandomNumberStr(n int) (randStr string) {
	charsLen := len(onlyNumber)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < n; i++ {
		randIndex := rand.Intn(charsLen)
		randStr += onlyNumber[randIndex : randIndex+1]
	}
	return randStr
}
