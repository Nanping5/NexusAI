package tts

import (
	"NexusAi/config"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	nls "github.com/aliyun/alibabacloud-nls-go-sdk"
)

// TTS Service 阿里云语音合成服务

type TTSService struct {
	token      string
	tokenMutex sync.RWMutex
	tokenExp   int64
	appkey     string
}

// TokenResult 阿里云Token响应
type TokenResult struct {
	ErrMsg string `json:"ErrMsg"`
	Token  struct {
		UserId     string `json:"UserId"`
		Id         string `json:"Id"`
		ExpireTime int64  `json:"ExpireTime"`
	} `json:"Token"`
}

var (
	ttsServiceOnce sync.Once
	ttsService     *TTSService
)

// GetTTSService 获取TTS服务单例
func GetTTSService() *TTSService {
	ttsServiceOnce.Do(func() {
		ttsService = &TTSService{}
	})
	return ttsService
}

// NewTTSService 创建TTS服务实例
func NewTTSService() *TTSService {
	return GetTTSService()
}

// GetToken 获取阿里云Token（带缓存）
func (s *TTSService) GetToken() (string, error) {
	s.tokenMutex.RLock()
	// 如果token存在且未过期（提前5分钟刷新）
	if s.token != "" && time.Now().Unix() < s.tokenExp-300 {
		token := s.token
		s.tokenMutex.RUnlock()
		return token, nil
	}
	s.tokenMutex.RUnlock()

	// 获取新token
	s.tokenMutex.Lock()
	defer s.tokenMutex.Unlock()

	// 双重检查
	if s.token != "" && time.Now().Unix() < s.tokenExp-300 {
		return s.token, nil
	}

	conf := config.GetConfig()
	akID := conf.VoiceServiceConfig.AccessKeyID
	akSecret := conf.VoiceServiceConfig.AccessKeySecret

	if akID == "" || akSecret == "" {
		return "", fmt.Errorf("aliyun AccessKeyID or AccessKeySecret not configured")
	}

	// 获取AppKey
	if s.appkey == "" {
		s.appkey = conf.VoiceServiceConfig.AppKey
	}

	// 使用阿里云SDK获取Token
	clientConfig := sdk.NewConfig()
	credential := credentials.NewAccessKeyCredential(akID, akSecret)
	client, err := sdk.NewClientWithOptions("cn-shanghai", clientConfig, credential)
	if err != nil {
		return "", fmt.Errorf("create client failed: %w", err)
	}

	request := requests.NewCommonRequest()
	request.Method = "POST"
	request.Domain = "nls-meta.cn-shanghai.aliyuncs.com"
	request.ApiName = "CreateToken"
	request.Version = "2019-02-28"

	response, err := client.ProcessCommonRequest(request)
	if err != nil {
		return "", fmt.Errorf("get token failed: %w", err)
	}

	// 解析响应
	var tr TokenResult
	if err := json.Unmarshal([]byte(response.GetHttpContentString()), &tr); err != nil {
		return "", fmt.Errorf("parse token response failed: %w", err)
	}

	if tr.ErrMsg != "" {
		return "", fmt.Errorf("get token error: %s", tr.ErrMsg)
	}

	s.token = tr.Token.Id
	s.tokenExp = tr.Token.ExpireTime

	return s.token, nil
}

// SetAppKey 设置AppKey
func (s *TTSService) SetAppKey(appkey string) {
	s.appkey = appkey
}

// TTSResult 语音合成结果
type TTSResult struct {
	AudioData []byte // 音频数据
	Format    string // 音频格式
}

// Synthesize 语音合成（同步方式）
func (s *TTSService) Synthesize(ctx context.Context, text string) ([]byte, error) {
	token, err := s.GetToken()
	if err != nil {
		return nil, fmt.Errorf("get token failed: %w", err)
	}

	if s.appkey == "" {
		return nil, fmt.Errorf("appkey not configured")
	}

	// 准备音频数据缓冲
	audioBuffer := &bytes.Buffer{}
	done := make(chan error, 1)

	// 创建日志（开启调试）
	logger := nls.NewNlsLogger(os.Stderr, "TTS", log.LstdFlags|log.Lmicroseconds)
	logger.SetLogSil(false) // 开启日志以便调试
	logger.SetDebug(true)

	// 连接配置
	connConfig := nls.NewConnectionConfigWithToken(nls.DEFAULT_URL, s.appkey, token)

	// 创建语音合成对象
	tts, err := nls.NewSpeechSynthesis(connConfig, logger, false,
		func(text string, param interface{}) {
			// onTaskFailed
			logger.Println("TaskFailed:", text)
			done <- fmt.Errorf("task failed: %s", text)
		},
		func(data []byte, param interface{}) {
			// onSynthesisResult - 接收音频数据
			n, err := audioBuffer.Write(data)
			logger.Printf("onSynthesisResult: received %d bytes, total %d, err: %v", len(data), n, err)
		},
		nil, // metainfo
		func(text string, param interface{}) {
			// onCompleted
			logger.Println("onCompleted:", text)
			done <- nil
		},
		func(param interface{}) {
			// onClose
			logger.Println("onClose: connection closed")
		},
		nil, // userParam
	)
	if err != nil {
		return nil, fmt.Errorf("create speech synthesis failed: %w", err)
	}

	// 设置合成参数（使用控制台配置：知小夏、16000Hz、WAV格式）
	param := nls.DefaultSpeechSynthesisParam()
	param.Voice = "zhixiaoxia" // 知小夏
	param.Format = nls.WAV
	param.SampleRate = 16000
	param.Volume = 50
	param.SpeechRate = 0
	param.PitchRate = 0

	logger.Println("Starting synthesis for text:", text)

	// 开始合成
	ch, err := tts.Start(text, param, nil)
	if err != nil {
		tts.Shutdown()
		return nil, fmt.Errorf("start synthesis failed: %w", err)
	}

	// 等待完成
	select {
	case err := <-done:
		tts.Shutdown()
		logger.Printf("Synthesis done, buffer size: %d bytes", audioBuffer.Len())
		if err != nil {
			return nil, err
		}
		return audioBuffer.Bytes(), nil
	case <-ch:
		tts.Shutdown()
		logger.Printf("Synthesis channel done, buffer size: %d bytes", audioBuffer.Len())
		return audioBuffer.Bytes(), nil
	case <-time.After(60 * time.Second):
		tts.Shutdown()
		return nil, errors.New("synthesis timeout")
	case <-ctx.Done():
		tts.Shutdown()
		return nil, ctx.Err()
	}
}

// SynthesizeAsync 异步语音合成，返回音频URL
func (s *TTSService) SynthesizeAsync(ctx context.Context, text string) (string, error) {
	// 阿里云实时语音合成是WebSocket方式，不直接返回URL
	// 如果需要URL方式，需要使用阿里云的异步长文本语音合成服务
	// 这里我们使用同步方式生成音频数据，然后保存到临时文件

	audioData, err := s.Synthesize(ctx, text)
	if err != nil {
		return "", fmt.Errorf("synthesize failed: %w", err)
	}

	// 保存到临时文件
	tmpFile, err := os.CreateTemp("", "tts-*.wav")
	if err != nil {
		return "", fmt.Errorf("create temp file failed: %w", err)
	}
	defer tmpFile.Close()

	if _, err := tmpFile.Write(audioData); err != nil {
		return "", fmt.Errorf("write audio data to file failed: %w", err)
	}

	return tmpFile.Name(), nil
}
