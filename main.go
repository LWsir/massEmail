package main

import (
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log"
	"math/rand"
	"net/http"
	"net/smtp"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func main() {
	// 创建 Gin 引擎
	r := gin.Default()

	r.LoadHTMLGlob("templates/*")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})

	r.GET("/download/:filename", func(c *gin.Context) {
		filename := c.Param("filename")
		file := "uploads/" + filename

		// 设置响应头，指定文件名和Content-Type
		c.Header("Content-Description", "File Transfer")
		c.Header("Content-Disposition", "attachment; filename="+filename)
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.File(file)
	})

	r.POST("send", func(c *gin.Context) {

		// 处理授权邮箱

		emailCountStr := c.PostForm("input-count")
		emailCount, err := strconv.Atoi(emailCountStr)
		if err != nil {
			log.Println("无效的邮箱数量")
			c.String(http.StatusBadRequest, "无效的邮箱数量")
			return
		}

		// 处理收件人、发送频率和发送内容
		intervalStr := c.PostForm("interval")
		interval, err := strconv.Atoi(intervalStr)
		if err != nil {
			log.Println("无效的发送评率")
			c.String(http.StatusBadRequest, "无效的发送评率")
			return
		}

		emails := make([]string, 0)
		passwords := make([]string, 0)
		errs := make([]string, 0)

		for i := 1; i <= emailCount; i++ {
			email := c.PostForm(fmt.Sprintf("email%d", i))
			password := c.PostForm(fmt.Sprintf("password%d", i))

			fmt.Println(email)
			trimmedStr := strings.TrimSpace(email)

			if trimmedStr != "" {
				emails = append(emails, email)
				passwords = append(passwords, password)
			}
			// 在这里可以对每个授权邮箱进行处理，例如存储到数据库或执行其他操作
			fmt.Printf("邮箱%d: %s\n", i, email)
			fmt.Printf("授权码%d: %s\n", i, password)
		}

		//处理发件人
		excelFile2, err2 := c.FormFile("excel2")
		if err2 != nil {
			fmt.Println(err2)
			if err2 != http.ErrMissingFile {
				c.String(400, "授权邮箱文件上传失败")
				return
			}
		} else {
			// 生成唯一的文件名
			fileName2 := uuid.New().String() + filepath.Ext(excelFile2.Filename)

			// 保存文件
			err4 := c.SaveUploadedFile(excelFile2, "uploads/"+fileName2)
			if err4 != nil {
				fmt.Println(err4)
				c.String(500, "文件保存失败")
				return
			}

			// 处理上传的文件
			log.Println("上传的文件名:", excelFile2.Filename)
			log.Println("保存的文件名:", fileName2)

			//处理数据
			f2, er2 := excelize.OpenFile("uploads/" + fileName2)
			if er2 != nil {
				fmt.Println("无法打开 Excel 文件:", er2)
				c.String(400, "无法打开 授权邮箱Excel 文件:")
				return
			}

			// 获取工作表名称列表
			sheetList2 := f2.GetSheetName(1)
			fmt.Printf(sheetList2)

			// 读取指定工作表的单元格数据
			rows2 := f2.GetRows(sheetList2)

			for _, row2 := range rows2 {
				// 获取Email和AuthCode列的数据
				email := strings.TrimSpace(row2[0])
				authCode := strings.TrimSpace(row2[1])

				// 打印Email和AuthCode
				fmt.Println("Email:", email)
				fmt.Println("AuthCode:", authCode)

				emails = append(emails, email)
				passwords = append(passwords, authCode)
			}
		}

		// 算出切片的长度
		len1 := len(emails)
		fmt.Println(len1)
		if len1 <= 0 {
			c.String(200, "授权邮箱为空请检查设置")
			return
		}

		content := c.PostForm("content")

		//处理收件人
		excelFile, err := c.FormFile("excel")

		//处理文件;
		if err != nil {
			log.Println(err)
			c.String(400, "收件人文件上传失败")
			return
		}

		// 生成唯一的文件名
		fileName := uuid.New().String() + filepath.Ext(excelFile.Filename)

		// 保存文件
		err = c.SaveUploadedFile(excelFile, "uploads/"+fileName)
		if err != nil {
			log.Println(err)
			c.String(500, "文件保存失败")
			return
		}

		// 处理上传的文件
		log.Println("上传的文件名:", excelFile.Filename)
		log.Println("保存的文件名:", fileName)

		//处理数据
		f, er := excelize.OpenFile("uploads/" + fileName)
		if er != nil {
			fmt.Println("无法打开 Excel 文件:", er)
			c.String(400, "无法打开 Excel 文件:")
			return
		}

		// 获取工作表名称列表
		sheetList := f.GetSheetName(1)
		fmt.Printf(sheetList)

		// 读取指定工作表的单元格数据
		rows := f.GetRows(sheetList)

		failCount := 0
		SuccessCount := 0

		//可用邮箱和授权码
		canUseEmails := make([]string, 0)
		canUsePasswords := make([]string, 0)

		//发送失败的邮箱
		failEmails := make([]string, 0)

		i := 0
		increment := true

		// 遍历行并输出单元格数据
		for _, row := range rows {
			for _, cell := range row {
				if cell == "" {

				} else {
					fmt.Printf("%s\t", cell)

					fmt.Printf("轮训到%d授权邮箱\n", i)
					fmt.Printf("延迟%d秒\n", interval)

					_, err1 := SendEmail(cell, emails[i], passwords[i], content)

					if err1 != nil {
						failCount++
						//str := err1.Error()
						//errMsg := ""
						//errMsg = getErrorMsg(str)
						//errs = append(errs, cell+" "+"发送失败："+errMsg+"\n")
						failEmails = append(failEmails, cell)
					} else {
						SuccessCount++
						errs = append(errs, cell+" "+"发送成功："+"\n")
						canUseEmails = append(canUseEmails, emails[i])
						canUsePasswords = append(canUsePasswords, passwords[i])
					}

					if increment {
						i++
						if i == len1 {
							increment = false
							i--
						}
					} else {
						i--
						if i == -1 {
							increment = true
							i = 0
						}
					}

					time.Sleep(time.Duration(interval) * time.Second)
				}
			}
		}

		fmt.Println("最开始的结果\n")
		fmt.Println(errs)

		length := len(failEmails)
		fmt.Println("最开始失败结果\n")
		fmt.Println(failEmails)

		if len(canUseEmails) <= 0 {
			// 构造弹框的JavaScript代码
			c.String(400, "所有授权邮箱都不可用")
			return
		}

		for a := 0; a < length; a++ {
			// 设置随机种子
			rand.Seed(time.Now().UnixNano())

			// 生成随机整数
			randomInt := rand.Intn(len(canUseEmails)) // 生成int范围内的随机整数

			_, err2 := SendEmail(failEmails[a], canUseEmails[randomInt], canUsePasswords[randomInt], content)

			if err2 != nil {
				errs = append(errs, failEmails[i]+" "+"发送失败："+getErrorMsg(err2.Error())+" "+"授权邮箱："+canUseEmails[randomInt]+"\n")
			} else {
				errs = append(errs, failEmails[i]+" "+"发送成功："+"\n")
				SuccessCount++
				failCount--
			}
		}

		allCount := failCount + SuccessCount

		//保存结果
		err4 := saveToExcel(errs)

		downloadLink := "/download/" + err4
		c.HTML(http.StatusOK, "res.html", gin.H{
			"res":          errs,
			"fail":         failCount,
			"success":      SuccessCount,
			"allCount":     allCount,
			"downloadLink": downloadLink,
		})

	})

	// 启动服务器
	r.Run(":8083")
}

/*
*
发送邮箱功能
*/
func SendEmail(str string, name string, password string, content string) (rtn int, err error) {
	fmt.Println(name)
	fmt.Println(password)
	fmt.Println(content)
	// 配置 SMTP 服务器信息
	smtpHost := "smtp.163.com"
	smtpPort := 25
	smtpUsername := name
	smtpPassword := password

	// 邮件内容
	from := name
	to := []string{str}
	subject := "最新邮件"
	body := content

	// 构建邮件内容
	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", from, to, subject, body)

	// 连接 SMTP 服务器
	fmt.Println("鉴权")
	auth := smtp.PlainAuth("", smtpUsername, smtpPassword, smtpHost)
	fmt.Println(auth)

	e := smtp.SendMail(fmt.Sprintf("%s:%d", smtpHost, smtpPort), auth, from, to, []byte(message))
	if e != nil {
		fmt.Println("邮件发送失败:", e)
		return 1, e
	} else {
		fmt.Println("邮件发送成功")
		return 0, nil
	}
}

func getErrorMsg(str string) (string3 string) {
	errMsg := ""
	switch {
	case strings.Contains(str, "500"):
		errMsg = "语法错误，请求的命令不符合163邮箱协议规范"
	case strings.Contains(str, "501"):
		errMsg = "参数错误，请求的命令参数不正确"
	case strings.Contains(str, "502"):
		errMsg = "命令不可实现，请求的命令无法被163邮箱服务器实现"
	case strings.Contains(str, "503"):
		errMsg = "命令序列错误，请求的命令序列顺序不正确"
	case strings.Contains(str, "504"):
		errMsg = "命令参数不支持，请求的命令参数不受163邮箱服务器支持"
	case strings.Contains(str, "535"):
		errMsg = "鉴权失败，邮箱和授权码不一致"
	case strings.Contains(str, "550"):
		errMsg = "无法发送邮件，表示163邮箱服务器拒绝了发送邮件的请求"
	case strings.Contains(str, "551"):
		errMsg = "用户不存在，表示163邮箱服务器无法找到收件人邮箱地址"
	case strings.Contains(str, "552"):
		errMsg = "存储空间不足，表示163邮箱服务器的存储空间已满或超过限制"
	case strings.Contains(str, "553"):
		errMsg = "无法发送邮件给指定的邮箱，表示163邮箱服务器不允许发送到该邮箱"
	default:
		errMsg = "未知错误"
	}
	return errMsg
}

func saveToExcel(data []string) (err3 string) {
	// 创建一个新的Excel文件
	f := excelize.NewFile()

	// 创建一个名为"Sheet1"的工作表
	index := f.NewSheet("Sheet1")

	// 将切片数据写入工作表中
	for i, value := range data {
		cell := fmt.Sprintf("A%d", i+1)
		f.SetCellValue("Sheet1", cell, value)
	}

	// 设置活动工作表
	f.SetActiveSheet(index)

	// 生成唯一的文件名
	fileName := uuid.New().String() + ".xlsx"

	// 将数据保存到文件
	if err := f.SaveAs("uploads/" + fileName); err != nil {
		fmt.Println("保存Excel文件失败:", err)
		return fileName
	} else {
		fmt.Println("Excel文件保存成功")
		return fileName
	}
}
