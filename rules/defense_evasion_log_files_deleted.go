// [metadata]
// creation_date = "2020/11/03"
// maturity = "production"
// updated_date = "2021/03/03"

// [rule]
// author = ["Elastic"]
// description = """
// Identifies the deletion of sensitive Linux system logs. This may indicate an attempt to evade detection or destroy
// forensic evidence on a system.
// """
// from = "now-9m"
// index = ["auditbeat-*", "logs-endpoint.events.*"]
// language = "eql"
// license = "Elastic License v2"
// name = "System Log File Deletion"
// references = [
//     "https://www.fireeye.com/blog/threat-research/2020/11/live-off-the-land-an-overview-of-unc1945.html",
// ]
// risk_score = 47
// rule_id = "aa895aea-b69c-4411-b110-8d7599634b30"
// severity = "medium"
// tags = ["Elastic", "Host", "Linux", "Threat Detection", "Defense Evasion"]
// timestamp_override = "event.ingested"
// type = "eql"

// query = '''
// file where event.type == "deletion" and
//   file.path :
//     (
//     "/var/run/utmp",
//     "/var/log/wtmp",
//     "/var/log/btmp",
//     "/var/log/lastlog",
//     "/var/log/faillog",
//     "/var/log/syslog",
//     "/var/log/messages",
//     "/var/log/secure",
//     "/var/log/auth.log"
//     )
// '''

// [[rule.threat]]
// framework = "MITRE ATT&CK"
// [[rule.threat.technique]]
// id = "T1070"
// name = "Indicator Removal on Host"
// reference = "https://attack.mitre.org/techniques/T1070/"

// [rule.threat.tactic]
// id = "TA0005"
// name = "Defense Evasion"
// reference = "https://attack.mitre.org/tactics/TA0005/"

package rules

import (
	"encoding/json"
	"log"
	"path/filepath"
	"strings"

	"github.com/mosajjal/ebpf-edr/types"
)

func defense_evasion_log_files_deleted() error {

	var suspicious_arguments = []string{
		"/var/run/utmp",
		"/var/log/wtmp",
		"/var/log/btmp",
		"/var/log/lastlog",
		"/var/log/faillog",
		"/var/log/syslog",
		"/var/log/messages",
		"/var/log/secure",
		"/var/log/auth.log",
	}

	var credSub types.EventSubscriber

	go func(s types.EventSubscriber) {
		log.Println("Running defense_evasion_log_files_deleted rule")
		s.Source = make(chan types.EventStream, 100)
		s.Subscribe()
		for {
			select {
			case event := <-s.Source:
				args_concat := event.Cmd + " " + strings.Join(event.Args, " ")
				if strings.Contains(args_concat, "rm") || strings.Contains(args_concat, "unlink") || strings.Contains(args_concat, "shred") {
					for _, argument := range event.Args {
						for _, arg := range suspicious_arguments {
							if m, _ := filepath.Match(arg, argument); m {
								event_json, _ := json.Marshal(event)
								log.Printf("Attempt to remove log files. Severity: High. Details: %s\n", string(event_json))
								break
							}
						}
					}
				}
			case <-types.GlobalQuit:
				return
				//todo:write quit
			}
		}
	}(credSub)
	return nil
}

var _ = defense_evasion_log_files_deleted()