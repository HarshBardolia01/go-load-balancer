package utils

import (
	"strconv"
)

func GetURL(schema string, ipAddress string, port int, endpoint string) string {
	return schema + "://" + ipAddress + ":" + strconv.Itoa(port) + "/" + endpoint
}
