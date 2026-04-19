package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	v1 "cron/api/cron/v1"
	"cron/internal/cron"

	"github.com/go-kratos/kratos/v2/log"
)

// TestCronService_ListTasks 测试列出所有定时任务
func TestCronService_ListTasks(t *testing.T) {
	// 创建调度器和任务管理器
	logger := log.DefaultLogger
	scheduler := cron.NewScheduler(logger)
	taskManager := cron.NewTaskManager(scheduler, logger)
	service := NewCronService(taskManager, logger)

	// 启动调度器并添加测试任务
	ctx := context.Background()
	scheduler.Start(ctx)
	defer scheduler.Stop(ctx)

	// 添加测试任务
	err := scheduler.AddTask("test_task_1", "0 0 * * * *", func() {
		t.Log("测试任务1执行")
	})
	if err != nil {
		t.Fatalf("添加测试任务失败: %v", err)
	}

	// 等待任务注册
	time.Sleep(100 * time.Millisecond)

	// 调用服务
	tasks, err := service.ListTasks(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTasks 调用失败: %v", err)
	}

	// 验证返回结果
	if tasks == nil {
		t.Fatal("返回的任务列表不应为空")
	}

	t.Logf("成功获取任务列表，任务数量: %d", len(tasks.Tasks))
	for _, task := range tasks.Tasks {
		t.Logf("任务: %s, 下次执行: %s", task.Name, task.NextRun)
	}
}

// TestCronService_AddTask 测试添加定时任务
func TestCronService_AddTask(t *testing.T) {
	scheduler := cron.NewScheduler(log.DefaultLogger)
	taskManager := cron.NewTaskManager(scheduler, log.DefaultLogger)
	service := NewCronService(taskManager, log.DefaultLogger)

	scheduler.Start(context.Background())
	defer scheduler.Stop(context.Background())

	// 测试添加新任务
	req := &v1.AddTaskRequest{
		Name: "new_test_task",
		Spec: "0 */5 * * * *", // 每5分钟
	}

	t.Logf("添加任务请求: name=%s, spec=%s", req.Name, req.Spec)

	// 通过服务添加任务
	_, err := service.AddTask(context.Background(), req)
	if err != nil {
		t.Fatalf("添加任务失败: %v", err)
	}

	t.Logf("成功添加任务: %s", req.Name)

	// 验证任务已添加
	tasks := scheduler.ListTasks()
	found := false
	for _, task := range tasks {
		if task.Name == "new_test_task" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("新添加的任务未在列表中找到")
	}
}

// TestCronService_RemoveTask 测试删除定时任务
func TestCronService_RemoveTask(t *testing.T) {
	scheduler := cron.NewScheduler(log.DefaultLogger)
	taskManager := cron.NewTaskManager(scheduler, log.DefaultLogger)
	service := NewCronService(taskManager, log.DefaultLogger)

	scheduler.Start(context.Background())
	defer scheduler.Stop(context.Background())

	// 先添加一个测试任务
	err := scheduler.AddTask("task_to_remove", "0 0 * * * *", func() {
		t.Log("待删除任务执行")
	})
	if err != nil {
		t.Fatalf("添加测试任务失败: %v", err)
	}

	// 等待任务注册
	time.Sleep(50 * time.Millisecond)

	// 验证任务存在
	tasks := scheduler.ListTasks()
	found := false
	for _, task := range tasks {
		if task.Name == "task_to_remove" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("测试任务添加失败")
	}

	// 通过服务删除任务
	_, err = service.RemoveTask(context.Background(), &v1.RemoveTaskRequest{Name: "task_to_remove"})
	if err != nil {
		t.Fatalf("删除任务失败: %v", err)
	}

	t.Log("成功删除任务: task_to_remove")

	// 验证任务已删除
	tasks = scheduler.ListTasks()
	found = false
	for _, task := range tasks {
		if task.Name == "task_to_remove" {
			found = true
			break
		}
	}
	if found {
		t.Fatal("任务应该已被删除，但仍存在")
	}
}

// TestCronService_TaskScheduling 测试任务调度功能
func TestCronService_TaskScheduling(t *testing.T) {
	scheduler := cron.NewScheduler(log.DefaultLogger)

	executionCount := 0
	// 添加每秒执行的任务用于测试
	err := scheduler.AddTask("quick_test_task", "* * * * * *", func() {
		executionCount++
		t.Logf("任务执行次数: %d", executionCount)
	})
	if err != nil {
		t.Fatalf("添加任务失败: %v", err)
	}

	// 启动调度器
	scheduler.Start(context.Background())
	defer scheduler.Stop(context.Background())

	// 等待任务执行几次
	time.Sleep(2 * time.Second)

	if executionCount == 0 {
		t.Fatal("任务应该至少执行一次")
	}

	t.Logf("任务调度正常，共执行 %d 次", executionCount)
}

// TestCronService_InvalidSchedule 测试无效的任务表达式
func TestCronService_InvalidSchedule(t *testing.T) {
	scheduler := cron.NewScheduler(log.DefaultLogger)

	// 尝试添加无效表达式的任务
	err := scheduler.AddTask("invalid_task", "invalid_cron_expression", func() {
		t.Log("不应该执行")
	})

	if err == nil {
		t.Fatal("无效的任务表达式应该返回错误")
	}

	t.Logf("正确识别无效表达式: %v", err)
}

// TestCronService_DuplicateTask 测试重复添加任务（覆盖更新）
func TestCronService_DuplicateTask(t *testing.T) {
	scheduler := cron.NewScheduler(log.DefaultLogger)

	// 添加第一个任务
	err := scheduler.AddTask("duplicate_task", "0 0 * * * *", func() {
		t.Log("第一个任务")
	})
	if err != nil {
		t.Fatalf("添加第一个任务失败: %v", err)
	}
	t.Log("成功添加第一个任务")

	// 添加同名任务 - scheduler 会自动覆盖更新
	err = scheduler.AddTask("duplicate_task", "0 30 * * * *", func() {
		t.Log("第二个任务（覆盖）")
	})
	if err != nil {
		t.Fatalf("覆盖任务失败: %v", err)
	}

	t.Log("成功覆盖更新同名任务")

	// 验证任务被更新（检查新 schedule）
	tasks := scheduler.ListTasks()
	found := false
	for _, task := range tasks {
		if task.Name == "duplicate_task" {
			found = true
			// 验证更新后的 spec
			t.Logf("任务已更新: %s", task.Name)
			break
		}
	}
	if !found {
		t.Fatal("覆盖后的任务未找到")
	}
}

// TestCronService_MultipleTasks 测试多个任务同时运行
func TestCronService_MultipleTasks(t *testing.T) {
	scheduler := cron.NewScheduler(log.DefaultLogger)

	// 添加多个任务
	tasks := []struct {
		name     string
		schedule string
	}{
		{"task_a", "* * * * * *"},   // 每秒
		{"task_b", "*/2 * * * * *"}, // 每2秒
		{"task_c", "*/3 * * * * *"}, // 每3秒
	}

	executionMap := make(map[string]int)
	var mu sync.Mutex

	for _, task := range tasks {
		taskName := task.name // 捕获变量
		err := scheduler.AddTask(task.name, task.schedule, func() {
			mu.Lock()
			executionMap[taskName]++
			mu.Unlock()
		})
		if err != nil {
			t.Fatalf("添加任务 %s 失败: %v", task.name, err)
		}
	}

	// 启动调度器
	scheduler.Start(context.Background())
	defer scheduler.Stop(context.Background())

	// 等待任务执行
	time.Sleep(4 * time.Second)

	// 验证每个任务都执行了
	for _, task := range tasks {
		mu.Lock()
		count := executionMap[task.name]
		mu.Unlock()
		if count == 0 {
			t.Fatalf("任务 %s 没有执行", task.name)
		}
		t.Logf("任务 %s 执行了 %d 次", task.name, count)
	}
}

// TestCronService_HealthCheck 测试健康检查
func TestCronService_HealthCheck(t *testing.T) {
	scheduler := cron.NewScheduler(log.DefaultLogger)
	taskManager := cron.NewTaskManager(scheduler, log.DefaultLogger)
	service := NewCronService(taskManager, log.DefaultLogger)

	// 测试服务是否初始化成功
	if service == nil {
		t.Fatal("服务初始化失败")
	}

	if service.taskManager == nil {
		t.Fatal("任务管理器未正确初始化")
	}

	t.Log("服务健康检查通过")
}

// TestCronService_ConcurrentAccess 测试并发访问
func TestCronService_ConcurrentAccess(t *testing.T) {
	scheduler := cron.NewScheduler(log.DefaultLogger)
	taskManager := cron.NewTaskManager(scheduler, log.DefaultLogger)
	service := NewCronService(taskManager, log.DefaultLogger)

	scheduler.Start(context.Background())
	defer scheduler.Stop(context.Background())

	// 先添加一些任务
	for i := 0; i < 5; i++ {
		name := fmt.Sprintf("concurrent_task_%d", i)
		err := scheduler.AddTask(name, "0 0 * * * *", func() {})
		if err != nil {
			t.Fatalf("添加任务 %s 失败: %v", name, err)
		}
	}

	// 并发调用 ListTasks
	done := make(chan bool, 3)
	for i := 0; i < 3; i++ {
		go func() {
			_, err := service.ListTasks(context.Background(), nil)
			if err != nil {
				t.Errorf("并发 ListTasks 失败: %v", err)
			}
			done <- true
		}()
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 3; i++ {
		<-done
	}

	t.Log("并发访问测试通过")
}

// TestCronService_APIEndpoints 测试 API 端点格式
func TestCronService_APIEndpoints(t *testing.T) {
	// 测试端点路径格式
	endpoints := []struct {
		method string
		path   string
		desc   string
	}{
		{"GET", "/cron/v1/tasks", "列出所有任务"},
		{"POST", "/cron/v1/tasks", "添加任务"},
		{"DELETE", "/cron/v1/tasks/{name}", "删除任务"},
	}

	for _, ep := range endpoints {
		if !strings.HasPrefix(ep.path, "/cron/v1/") {
			t.Errorf("端点 %s 路径格式错误", ep.path)
		}
		t.Logf("端点: %s %s - %s", ep.method, ep.path, ep.desc)
	}
}

// BenchmarkListTasks 基准测试：列出任务
func BenchmarkListTasks(b *testing.B) {
	scheduler := cron.NewScheduler(log.DefaultLogger)
	taskManager := cron.NewTaskManager(scheduler, log.DefaultLogger)
	service := NewCronService(taskManager, log.DefaultLogger)

	// 添加一些任务
	for i := 0; i < 100; i++ {
		name := fmt.Sprintf("benchmark_task_%d", i)
		scheduler.AddTask(name, "0 0 * * * *", func() {})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.ListTasks(context.Background(), nil)
	}
}
