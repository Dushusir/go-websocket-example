package utils

import (
	"math/rand"
	"time"
)

func GenerateID(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var result []byte
	for i := 0; i < length; i++ {
		result = append(result, charset[rand.Intn(len(charset))])
	}
	return string(result)
}

// 用于生成随机数种子
var r = rand.New(rand.NewSource(time.Now().UnixNano()))

// 当前随机生成的人名
var prevName *Name

// 人名结构体
type Name struct {
	First string // 名字
}

// 随机生成英文名字函数
func generateName() *Name {
	

	// 名字列表
	firstNames := []string{"Emma", "Noah", "Olivia", "Liam", "Ava", "William", "Sophia", "Mason", "Isabella", "James", "Mia", "Benjamin", "Charlotte", "Jacob", "Amelia"}

	// 随机选择一个名字
	firstName := firstNames[r.Intn(len(firstNames))]

	return &Name{First: firstName}
}

// 生成不重复的英文名字函数
func generateUniqueName(prevName *Name) *Name {
	for {
		// 生成一个名字
		name := generateName()

		// 如果这个名字和上一个名字不相同，返回这个名字
		if prevName == nil || name.First != prevName.First {
			return name
		}
	}
}

func GetUniqueName() *Name {
	name := generateUniqueName(prevName)

	prevName = name

	return name

}
