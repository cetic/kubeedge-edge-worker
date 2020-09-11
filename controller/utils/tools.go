package utils

import(
  "time"
  "strconv"
)

func GetTimeStamp() string {
  return strconv.FormatInt(time.Now().Unix(), 10)
}
