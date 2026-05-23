package server

import "fmt"

func FormatAlertText(alert AlertDetail) string {
	statusIcon := "🔴"
	if alert.Status == "resolved" {
		statusIcon = "🟢"
	}
	return fmt.Sprintf("🚨 ALERT\n\n%s %s: %s", statusIcon, alert.Labels["alertname"], alert.Annotations["summary"])
}
