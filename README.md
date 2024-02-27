# csvutil

대용량 데이터를 csv파일 형태로 다운로드하는 라이브러리 <br>

### Requirement
***
* 라이브러리
    *  gin-gonic : https://github.com/gin-gonic/gin
    * gorm : https://github.com/go-gorm/gorm
* 데이터베이스
    * postgresql


### Installation
***
```
go get github.com/innodep-tms/csvutil
```

### Usage
***
#### 일반 사용 예시 (gin-gonic handler)
* csv 태그를 이용하여 테이블 라벨을 지정할 수 있다.
```
package handler

import (
    "fmt"

    "github.com/gin-gonic/gin"
	"gorm.io/gorm"
    "github.com/innodep-tms/csvutil"
)

func HandlerFunc(c *gin.Context) {
    db, _ := gorm.Open()

    err := csvutil.TransferCSVFileChunked[data](c, db, "query", "filename", 1000)

    if err != nil {
        fmt.Println(err)
    }
}

type data struct {
    Name string `gorm:"column:name" csv:"이름"
    Value int `gorm:"column:value" csv:"값"
}
```


