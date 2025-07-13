package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
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

func main() {
	// 显示启动横幅
	showBanner()

	// 解析命令行参数
	configPath := flag.String("config", "./config.yaml", "配置文件路径")
	flag.Parse()

	fmt.Printf("🔧 正在加载配置文件: %s\n", *configPath)

	// 加载配置文件，支持自动创建
	cfg, created, err := config.LoadConfigWithAutoCreate(*configPath)
	if err != nil {
		log.Fatalf("❌ 配置文件处理失败: %v", err)
	}

	// 如果创建了新的配置文件，提示用户并退出
	if created {
		fmt.Printf("\n✨ 已创建默认配置文件: %s\n", *configPath)
		fmt.Println("📝 请根据您的环境修改配置文件中的以下关键设置：")
		fmt.Println("   • grunichat.url: GRUniChat 服务器地址")
		fmt.Println("   • grunichat.client_id: 客户端标识（建议改为有意义的名称）")
		fmt.Println("   • onebot.websocket_url: OneBot 服务器地址")
		fmt.Println("   • filter.service_groups: 服务的QQ群列表")
		fmt.Println()
		fmt.Println("⏰ 程序将在 5 秒后退出，请修改配置文件后重新启动...")

		// 倒计时
		for i := 5; i > 0; i-- {
			fmt.Printf("\r⏳ %d 秒后退出...", i)
			time.Sleep(1 * time.Second)
		}
		fmt.Println()
		os.Exit(0)
	}

	fmt.Println("✅ 配置文件加载成功")

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

	fmt.Println("🚀 正在启动 GRUniChat-OneBot 模块化适配器...")
	logger.Info("Starting GRUniChat-OneBot Modular Adapter")

	// 创建并启动模块化适配器
	ctx := context.Background()
	adapterInstance := adapter.NewModularAdapter(cfg, logger)
	if err := adapterInstance.Start(ctx); err != nil {
		logger.Fatalf("Failed to start adapter: %v", err)
	}
}
