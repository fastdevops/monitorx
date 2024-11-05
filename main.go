package main

import (
	"github.com/fastdevops/monitorx/config"
	"github.com/fastdevops/monitorx/global"
	"github.com/fastdevops/monitorx/logger"
	"github.com/fastdevops/monitorx/monitor"
	"github.com/fastdevops/monitorx/router"
	"github.com/fastdevops/monitorx/utils"

	"strconv"
)

func main() {
	// Initialize configuration
	cf := config.InitConfigSet()
	global.Config = cf

	// Initialize logging
	logger.InitLoggerConfig(cf.Logger.LogLevel, cf.Logger.LogPath, cf.Logger.LogName, cf.Logger.LogMaxSize, cf.Logger.LogMaxBackups, cf.Logger.LogMaxAge)
	logger.Logger.Info(`
######################################################################################
==========================     欢迎使用大数据集群监控平台     ==========================
==========================          运维团队出品             ==========================
######################################################################################
	 _     _           _       _                                 _ _
	| |__ (_) __ _  __| | __ _| |_ __ _   _ __ ___   ___  _ __  (_) |_ ___  _ __
	| '_ \| |/ _' |/ _' |/ _' | __/ _' | | '_ ' _ \ / _ \| '_ \ | | __/ _ \| '__|
	| |_) | | (_| | (_| | (_| | || (_| | | | | | | | (_) | | | || | || (_) | |
	|_.__/|_|\__, |\__,_|\__,_|\__\__,_| |_| |_| |_|\___/|_| |_|/ |\__\___/|_|
	         |___/
	`)
	// Convert check time
	checks_time_second, err := strconv.Atoi(cf.Check.MAX_SECOND)
	if err != nil {
		logger.Logger.Error("Invalid check time configuration.")
		return
	}

	// Run github.com/fastdevops/monitorx
	hadoopStatus := &utils.HadoopCluster{}
	go monitor.RunCheck(cf, checks_time_second, hadoopStatus)

	// Initialize router with pongo2 templates
	router.InitRouterConfig(cf)
}
