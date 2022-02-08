package service

func getQueryItemsStartAndEnd(pageNum, pageSize, listLen int) (int, int) {
	start := (pageNum - 1) * pageSize
	end := pageNum*pageSize - 1
	if listLen <= pageNum {
		start, end = 0, listLen
	}
	if start > listLen-1 {
		start, end = 0, pageSize-1
	}
	if end > listLen {
		end = listLen
	}
	return start, end
}
