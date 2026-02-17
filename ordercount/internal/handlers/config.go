package handlers

// WecomWebhook 用于存放企业微信机器人 webhook URL，由 main 在启动时从配置文件注入。
var WecomWebhook string

// DoubaoAI 相关配置，由 main 在启动时从配置文件或环境变量注入。
var DoubaoAPIKey string
var DoubaoEndpoint string
var DoubaoModel string
