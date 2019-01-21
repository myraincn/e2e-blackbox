/**
*  @file
*  @copyright defined in go-seele/LICENSE
 */

package main

import (
	"encoding/json"
	"fmt"
	"net/mail"
	"net/smtp"
	"os/exec"
	"strings"
	"time"

	"github.com/scorredoira/email"
	"github.com/seeleteam/e2e-blackbox/store"
)

var (
	attachFile = []string{}
)

// config.go
const (
	Path          = "github.com/seeleteam/go-seele/e2e-blackbox"
	CoverFileName = "seele_coverage_detail"
	CoverPackage  = "common\t,core\t,trie\t,p2p\t,seele\t,client\t,contract\t,domain\t,HTLC\t,light\t,network\t,subchain\t,transfer\t"

	Subject    = "Daily Blackbox E2E Test Report"
	Sender     = "send@email.com"
	Password   = "password"
	SenderName = "reporter"
	CC         = "CC@email.com"
	Receivers  = "receiver@email.com"
	Host       = "smtp-mail.outlook.com:587"
	//Host       = "smtp.exmail.qq.com:25"

	StartHour = 04
	StartMin  = 00
	StartSec  = 00

	BenchTopN         = "15"
	BenchReportFormat = "pdf"
)

func main() {
	now := time.Now()
	weekday := now.Weekday()
	if weekday != time.Saturday && weekday != time.Sunday {
		fmt.Println("Go")
		do(now.Format("20060102"))
	}
}

func sendEmail(message string, attachFile []string) {
	// fmt.Println(message, attachFile)
	msg := email.NewMessage(Subject, message)
	msg.From, msg.To = mail.Address{Name: SenderName, Address: Sender}, strings.Split(Receivers, ";")
	msg.Cc = strings.Split(CC, ";")
	for _, filePath := range attachFile {
		if err := msg.Attach(filePath); err != nil {
			fmt.Printf("failed to add attach file. path: %s, err: %s\n", filePath, err)
		}
	}

	hp := strings.Split(Host, ":")
	auth := smtp.PlainAuth("", Sender, Password, hp[0])

	if err := email.Send(Host, auth, msg); err != nil {
		fmt.Println("failed to send mail. err:", err)
	}
}

func do(today string) {
	coverResult, specified := Run()
	coverbyte, err := json.Marshal(specified)
	if err != nil {
		fmt.Println("Marshal specified FAIL")
	}
	fmt.Println("cover done")
	// save the result
	store.Save(today, coverbyte)
	fmt.Println("saved data")
	message := ""
	if strings.Contains(coverResult, "FAIL") {
		message += "😦 😦 😦 discover bug!\n\n"
	} else {
		message += "😁 Good day with no error~\n\n"
		// attachFile = append(attachFile, CoverFileName+".html")
	}

	// message += PrintSpecifiedPkg(yesterday, specified)
	message += "\n\n============= Go cover seele cmd commands completed. ===============\n" + coverResult

	sendEmail(message, attachFile)
}

// Run go
func Run() (all string, specified map[string]string) {
	specified = make(map[string]string)

	// coverbyte, err := exec.Command("go", "test", "./...", "-v", "-timeout", "3h", "-coverprofile="+CoverFileName).CombinedOutput()
	coverbyte, err := exec.Command("go", "test", "./...", "-v", "-timeout", "3h").CombinedOutput()
	if err != nil {
		return fmt.Sprintf("cover FAIL: %s %s", err, string(coverbyte)), nil
	}

	// remove useless output
	outs, pkgs := strings.Split(string(coverbyte), "\n"), strings.Split(CoverPackage, ",")
	for _, out := range outs {
		// ? == 63
		if out == "" || out[0] == 63 {
			continue
		}

		for _, pkg := range pkgs {
			if strings.Contains(out, pkg) {
				specified[pkg] = out
			}
		}

		all += out + "\n"
	}

	// go tool cover -html=covprofile -o coverage.html
	// if err := exec.Command("go", "tool", "cover", "-html="+CoverFileName, "-o", CoverFileName+".html").Run(); err != nil {
	// 	return fmt.Sprintf("tool cover FAIL: %s", err), nil
	// }

	return all, specified
}

// PrintSpecifiedPkg print pkg
func PrintSpecifiedPkg(yestoday string, specified map[string]string) string {
	result := "\n============= Change in coverage of major packages compared to yesterday ===============\n\n"
	yestodaySpec := make(map[string]string)
	coverByte := store.Get(yestoday)
	if err := json.Unmarshal(coverByte, &yestodaySpec); err != nil {
		return ""
	}

	for k, v := range specified {
		out, ok := yestodaySpec[k]
		if !ok {
			result += v + "\n"
		} else {
			result += out + " --> " + v[strings.Index(v, "coverage"):] + "\n"
		}
	}

	return result
}
