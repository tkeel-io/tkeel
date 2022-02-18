package service

import (
	"strings"

	"github.com/tkeel-io/tkeel/pkg/model"
)

func getReglarStringKeyWords(keyWords string) string {
	words := strings.TrimSpace(keyWords)
	if words == "" {
		return ".*"
	}
	return ".*" + words + ".*"
}

func getQueryItemsStartAndEnd(pageNum, pageSize, listLen int) (int, int) {
	if pageSize == 0 && pageNum == 0 {
		return 0, listLen
	}
	if pageSize < 0 {
		pageSize = 0
	}
	if pageNum < 0 {
		pageNum = 0
	}
	start := (pageNum - 1) * pageSize
	end := pageNum*pageSize - 1
	if listLen <= pageSize {
		start, end = 0, listLen
	}
	if start > listLen-1 {
		start, end = listLen-1, listLen
	}
	if end > listLen {
		end = listLen
	}
	if end < 0 {
		end = 0
	}
	if start < 0 {
		start = 0
	}
	return start, end
}

func pluginIsTkeelComponent(pluginID string) bool {
	for _, v := range model.TKeelComponents {
		if v == pluginID {
			return true
		}
	}
	return false
}
