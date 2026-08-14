package snowflake

import (
	"sync"
	"testing"

	"shorturl/internal/pkg/base62"
)

func TestSnowflakeAndBase62(t *testing.T) {
	// 1. 初始化雪花算法（设置机器编号为 1）
	err := Init("2026-01-01", 1)
	if err != nil {
		t.Fatalf("雪花算法初始化失败: %v", err)
	}

	// 2. 并发生成 10 个 ID 并转成短码打印出来
	t.Log("=========== 生成短链接演示 ===========")
	for i := 0; i < 10; i++ {
		id := GenID()                     // ① 生成雪花 ID
		code := base62.Encode(uint64(id)) // ② 转成 62 进制短码
		t.Logf("雪花 ID: %d  ➔  短链接后缀: %s (http://s.cn/%s)", id, code, code)
	}

	// 3. 高并发防重测试：并发生成 10000 个 ID，验证绝无重复！
	var wg sync.WaitGroup
	var idMap sync.Map
	count := 10000

	wg.Add(count)
	for i := 0; i < count; i++ {
		go func() {
			defer wg.Done()
			id := GenID()
			// 如果 map 里已经存在相同的 ID，说明有碰撞重复！
			if _, loaded := idMap.LoadOrStore(id, true); loaded {
				t.Errorf("发现重复的 ID: %d", id)
			}
		}()
	}
	wg.Wait()
	t.Logf("✅ 高并发测试通过：瞬间并发生成 %d 个 ID，0 重复、0 碰撞！", count)
}
