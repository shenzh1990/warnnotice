package util

import (
	"fmt"
	"github.com/gotoeasy/glang/cmn"
	"log"
	"os"
	"strconv"
	"time"
)

var DoDebug = false
var logChan = make(chan cmn.GlcData, 10000)

var localCount = 0
var localFile *os.File
var localLogger *log.Logger

const logRoot = "./log/"
const LogLevelError = "error"
const LogLevelInfo = "info"
const LogLevelDebug = "debug"

func initLog(path string) (*log.Logger, *os.File, error) {
	//removeOldFiles()
	_, err := os.Stat(logRoot)
	if err != nil && os.IsNotExist(err) {
		_ = os.Mkdir(logRoot, os.ModePerm)
	}
	file, errFile := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
	if errFile != nil {
		return nil, nil, errFile
	}
	return log.New(file, "", log.Ldate|log.Ltime|log.Ldate|log.Lmicroseconds), file, nil
}

func getPath(name string) string {
	year, month, date := GetNow().Date()
	strDate := strconv.Itoa(year) + month.String() + strconv.Itoa(date)
	strTime := strconv.Itoa(GetNow().Hour()) + "_" + strconv.Itoa(GetNow().Minute())
	return logRoot + name + strDate + "_" + strTime + ".log"
}

func LogChanHandler(apiUrl, systemName string) {
	cmn.SetGlcClient(cmn.NewGlcClient(&cmn.GlcOptions{
		ApiUrl:           apiUrl,
		Enable:           "true",
		EnableConsoleLog: "true",
		System:           systemName,
	}))
	for {
		glcData := <-logChan
		if apiUrl == `local` {
			if localCount%10000 == 0 {
				if localFile != nil {
					_ = localFile.Close()
				}
				localLogger, localFile, _ = initLog(getPath("local"))
			}
			localCount++
			localLogger.Println(glcData.Text)
		} else {
			if float64(len(logChan))/float64(cap(logChan)) > 0.9 {
				continue
			} else if len(logChan) == cap(logChan)/2 {
				cmn.Error(cmn.GlcData{Text: fmt.Sprintf(`log chan %d`, cap(logChan)/2), LogLevel: `error`})
			}
			switch glcData.LogLevel {
			case LogLevelError:
				cmn.Error(glcData)
			case LogLevelInfo:
				cmn.Info(glcData)
			case LogLevelDebug:
				cmn.Debug(glcData)
			}
		}

	}
}

func Log(logLevel, content string) {
	glcData := cmn.GlcData{Text: content, LogLevel: logLevel}
	logChan <- glcData
}

func InfoSync(msg string) {
	logContent := &cmn.GlcData{Text: msg, System: "infoSync"}
	cmn.Info(logContent)
}

func LogLess(logLevel, content string) {
	if time.Now().Second() == 0 {
		Log(logLevel, content)
	}
}
