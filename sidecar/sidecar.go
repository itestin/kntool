package sidecar

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

// HandlerLatency handler latency
func HandlerLatency(c *gin.Context) {
	val, ok := c.Params.Get("latency")
	if !ok {
		c.JSON(400, gin.H{
			"error": "required params of latency",
		})
	}

	ctx, cancel := context.WithTimeout(c, time.Second*10)
	defer cancel()

	log, err := delay(ctx, val)
	c.JSON(200, gin.H{
		"error": err,
		"log":   log,
	})
}

// run tc command
func run(ctx context.Context, args []string) (string, error) {
	environments := make([]string, 0)
	for _, e := range os.Environ() {
		environments = append(environments, e)
	}

	sysProcAttr := &syscall.SysProcAttr{
		Setpgid: true, // 使子进程拥有自己的 pgid，等同于子进程的 pid
	}

	cmd := exec.CommandContext(ctx, "tc", args...)

	log := bytes.NewBufferString("")
	cmd.Stdout = log
	cmd.Stderr = log
	cmd.SysProcAttr = sysProcAttr
	cmd.Env = environments
	if err := cmd.Start(); err != nil {
		return log.String(), err
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- cmd.Wait()
	}()

	// 可以通过 context 控制命令执行，调用方可以调用 cancel 或者设置超时控制命令执行生命周期
	// 如果进程执行失败，应当 kill 整个进程组，防止该进程 fork 的子进程逃逸
	done := ctx.Done()
	for {
		select {
		case <-done:
			done = nil
			pid := cmd.Process.Pid
			if err := syscall.Kill(-1*pid, syscall.SIGKILL); err != nil {
				return log.String(), err
			}
		case err := <-errCh:
			if done == nil {
				return log.String(), ctx.Err()
			}
			return log.String(), err
		}
	}
}

func show() {

}

func delay(ctx context.Context, value string) (string, error) {
	return run(ctx, []string{"qdisc", "add", "dev", "eth0", "root", "netem", "delay", value})
}

func duplicate() {

}

func loss() {

}
