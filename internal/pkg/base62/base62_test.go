package base62

import (
	"testing"
)

func TestBase62(t *testing.T) {
	// 测试用例：给定几个数字，测试转成短码再还原是否一致
	testNumbers := []uint64{0, 1, 62, 63, 1000, 123456789, 9876543210}

	for _, num := range testNumbers {
		// 1. 数字转成短码
		code := Encode(num)

		// 2. 短码还原回数字
		decodedNum, err := Decode(code)
		if err != nil {
			t.Fatalf("解码失败: %v", err)
		}

		t.Logf("原始数字: %10d ➔ 62进制短码: %-8s ➔ 还原数字: %d", num, code, decodedNum)

		// 3. 断言还原出来的数字必须和原始数字完全一致
		if num != decodedNum {
			t.Errorf("期望数字 %d, 但还原得到 %d", num, decodedNum)
		}
	}
}
