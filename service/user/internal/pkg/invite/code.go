package invite

import (
	"crypto/rand"
	"math/big"
)

//	var AlphanumericSet = []rune{
//		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
//		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
//		'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
//	}
//
//	func GetInvCodeByUID(uid uint64, l int) string {
//		var code []rune
//		slIdx := make([]byte, l)
//		for i := 0; i < l; i++ {
//			slIdx[i] = byte(uid % uint64(len(AlphanumericSet)))               // 获取 62 进制的每一位值
//			idx := (slIdx[i] + byte(i)*slIdx[0]) % byte(len(AlphanumericSet)) // 其他位与个位加和再取余（让个位的变化影响到所有位）
//			code = append(code, AlphanumericSet[idx])
//			uid = uid / uint64(len(AlphanumericSet)) // 相当于右移一位（62进制）
//		}
//		return string(code)
//	}
const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

func GenerateInviteCode(length int) (string, error) {
	b := make([]byte, length)
	charsetSize := big.NewInt(int64(len(charset)))

	for i := range b {
		index, err := rand.Int(rand.Reader, charsetSize)
		if err != nil {
			return "", err
		}
		b[i] = charset[index.Int64()]
	}
	return string(b), nil
}
