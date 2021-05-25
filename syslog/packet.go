/*
Copyright (c) 2013 Paul Morton, Papertrail, Inc., & Paul Hammond

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package syslog

import (
	"fmt"
	"strings"
	"time"
)

// A Packet represents an RFC5424 syslog message
type Packet struct {
	Severity       Priority
	Facility       Priority
	StructuredData string
	Hostname       string
	Tag            string
	Time           time.Time
	Message        string
}

// like time.RFC3339Nano but with a limit of 6 digits in the SECFRAC part
const rfc5424time = "2006-01-02T15:04:05.999999Z07:00"

// The combined Facility and Severity of this packet. See RFC5424 for details.
func (p Packet) Priority() Priority {
	return (p.Facility << 3) | p.Severity
}

func (p Packet) cleanMessage() string {
	s := strings.Replace(p.Message, "\n", " ", -1)
	s = strings.Replace(s, "\r", " ", -1)
	return strings.Replace(s, "\x00", " ", -1)
}

func (p Packet) structuredData() string {
	if p.StructuredData == "" {
		return "-"
	}
	return p.StructuredData
}

// Generate creates a RFC5424 syslog format string for this packet.
func (p Packet) Generate(max_size int) string {
	ts := p.Time.Format(rfc5424time)
	if max_size == 0 {
		return fmt.Sprintf("<%d>1 %s %s %s rs2 - %s %s", p.Priority(), ts, p.Hostname, p.Tag, p.structuredData(), p.cleanMessage())
	} else {
		msg := fmt.Sprintf("<%d>1 %s %s %s rs2 - %s %s", p.Priority(), ts, p.Hostname, p.Tag, p.structuredData(), p.cleanMessage())
		if len(msg) > max_size {
			return msg[0:max_size]
		} else {
			return msg
		}
	}
}

// A convenience function for testing
func Parse(line string) (Packet, error) {
	var (
		packet   Packet
		priority int
		ts       string
		hostname string
		tag      string
	)

	splitLine := strings.Split(line, " rs2 - - ")
	if len(splitLine) != 2 {
		return packet, fmt.Errorf("couldn't parse %s", line)
	}

	fmt.Sscanf(splitLine[0], "<%d>1 %s %s %s", &priority, &ts, &hostname, &tag)

	t, err := time.Parse(rfc5424time, ts)
	if err != nil {
		return packet, err
	}

	return Packet{
		Severity: Priority(priority & 7),
		Facility: Priority(priority >> 3),
		Hostname: hostname,
		Tag:      tag,
		Time:     t,
		Message:  splitLine[1],
	}, nil
}
