package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"grunichat-onebot-adapter/internal/adapter"
	"grunichat-onebot-adapter/internal/config"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

// 显示启动横幅
func showBanner() {
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("    ____ ____  _   _       _  ____ _           _                               ")
	fmt.Println("   / ___|  _ \\| | | |_ __ (_)/ ___| |__   __ _| |_                             ")
	fmt.Println("  | |  _| |_) | | | | '_ \\| | |   | '_ \\ / _` | __|                            ")
	fmt.Println("  | |_| |  _ <| |_| | | | | | |___| | | | (_| | |_                             ")
	fmt.Println("   \\____|_| \\_\\\\___/|_| |_|_|\\____|_| |_|\\__,_|\\__|    _           _   _       ")
	fmt.Println("                          / _ \\ _ __   ___| |__   ___ | |_  __   _/ | / |      ")
	fmt.Println("                         | | | | '_ \\ / _ \\ '_ \\ / _ \\| __| \\ \\ / / | | |      ")
	fmt.Println("                         | |_| | | | |  __/ |_) | (_) | |_   \\ V /| | | |      ")
	fmt.Println("                          \\___/|_| |_|\\___|_.__/ \\___/ \\__|   \\_/ |_| |_|      ")
	fmt.Println()
	fmt.Printf("                             WebSocket 消息广播器 %s                          \n", Version)
	fmt.Println()
	fmt.Println("                    作者: Glory Redstone Union - caikun233                          ")
	fmt.Println("                     描述: GRuniChat协议与Onebot协议适配器              ")
	if BuildTime != "unknown" {
		fmt.Printf("                         编译时间: %s                        \n", BuildTime)
	}
	fmt.Println("═══════════════════════════════════════════════════════════════════════════════")
	fmt.Println()
}

// GitHub Release API 响应结构
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	HTMLURL string `json:"html_url"`
}

// 检查版本更新
func checkForUpdates() {
	fmt.Print("正在检查版本更新...")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/your-org/grunichat-onebot/releases/latest")
	if err != nil {
		fmt.Println(" 无法检查更新")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Println(" 无法获取版本信息")
		return
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		fmt.Println(" 解析版本信息失败")
		return
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentVersion := strings.TrimPrefix(Version, "v")

	if latestVersion == currentVersion || Version == "dev" {
		fmt.Println(" 已是最新版本")
	} else {
		fmt.Printf(" 发现新版本！\n")
		fmt.Printf("   当前版本: %s\n", currentVersion)
		fmt.Printf("   最新版本: %s\n", latestVersion)
		fmt.Printf("   下载地址: %s\n", release.HTMLURL)
		fmt.Println()
	}
}

func main() {
	// 显示启动横幅
	showBanner()

	// 解析命令行参数
	configPath := flag.String("config", "./config.yaml", "配置文件路径")
	noCheckUpdate := flag.Bool("no-check-update", false, "跳过版本更新检查")
	flag.Parse()

	// 检查版本更新（除非用户明确跳过）
	if !*noCheckUpdate {
		checkForUpdates()
	}

	fmt.Printf("正在加载配置文件: %s\n", *configPath)

	// 加载配置文件，支持自动创建
	cfg, created, err := config.LoadConfigWithAutoCreate(*configPath)
	if err != nil {
		log.Fatalf("配置文件处理失败: %v", err)
	}

	// 如果创建了新的配置文件，提示用户并退出
	if created {
		fmt.Printf("\n已创建默认配置文件: %s\n", *configPath)
		fmt.Println("请根据您的环境修改配置文件中的以下关键设置：")
		fmt.Println("   • grunichat.url: GRUniChat 服务器地址")
		fmt.Println("   • grunichat.client_id: 客户端标识（建议改为有意义的名称）")
		fmt.Println("   • onebot.websocket_url: OneBot 服务器地址")
		fmt.Println("   • filter.service_groups: 服务的QQ群列表")
		fmt.Println()
		fmt.Println("程序将在 5 秒后退出，请修改配置文件后重新启动...")

		// 倒计时
		for i := 5; i > 0; i-- {
			fmt.Printf("\r%d 秒后退出...", i)
			time.Sleep(1 * time.Second)
		}
		fmt.Println()
		os.Exit(0)
	}

	fmt.Println("配置文件加载成功")

	// 设置日志
	logger := logrus.New()

	// 设置日志级别
	level, err := logrus.ParseLevel(cfg.Log.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// 设置日志格式
	if cfg.Log.Format == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	// 设置日志输出
	if cfg.Log.File != "" {
		file, err := os.OpenFile(cfg.Log.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			logger.SetOutput(file)
		} else {
			logger.Warnf("Failed to open log file %s: %v", cfg.Log.File, err)
		}
	}

	fmt.Println("正在启动 GRUniChat-OneBot 模块化适配器...")
	logger.Info("Starting GRUniChat-OneBot Modular Adapter")

	// 创建并启动模块化适配器
	ctx := context.Background()
	adapterInstance := adapter.NewModularAdapter(cfg, logger)
	if err := adapterInstance.Start(ctx); err != nil {
		logger.Fatalf("Failed to start adapter: %v", err)
	}

	// 检查版本更新
	checkForUpdates()
}
