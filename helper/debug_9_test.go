/*
 * Copyright 2018-2020 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package helper_test

import (
	"log"
	"os"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/libjvm/helper"
	"github.com/sclevine/spec"
)

func testDebug9(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
		d      = helper.Debug9{}
	)

	it("does nothing if $BPL_DEBUG_ENABLED is no set", func() {
		Expect(d.Execute()).To(BeNil())
	})

	context("$BPL_DEBUG_ENABLED", func() {
		var fakeIPv6File *os.File
		it.Before(func() {
			Expect(os.Setenv("BPL_DEBUG_ENABLED", "true")).
				To(Succeed())

			var fakeIPv6FileErr error
			fakeIPv6File, fakeIPv6FileErr = os.CreateTemp("", "IPv6Test")
			if fakeIPv6FileErr != nil {
				log.Fatal(fakeIPv6FileErr)
			}
			fakeIPv6File.WriteString("0")
			d.CustomIPv6CheckPath = fakeIPv6File.Name()
		})

		it.After(func() {
			Expect(os.Unsetenv("BPL_DEBUG_ENABLED")).To(Succeed())
			os.Remove(fakeIPv6File.Name())
		})

		it("contributes configuration", func() {
			Expect(d.Execute()).To(Equal(map[string]string{
				"JAVA_TOOL_OPTIONS": "-agentlib:jdwp=transport=dt_socket,server=y,address=*:8000,suspend=n",
			}))
		})

		context("jdwp agent already configured", func() {
			it.Before(func() {
				Expect(os.Setenv("JAVA_TOOL_OPTIONS", "-agentlib:jdwp=something")).To(Succeed())
			})

			it.After(func() {
				Expect(os.Unsetenv("JAVA_TOOL_OPTIONS")).To(Succeed())
			})

			it("does not update JAVA_TOOL_OPTIONS", func() {
				Expect(d.Execute()).To(BeEmpty())
			})
		})

		context("$BPL_DEBUG_PORT", func() {
			it.Before(func() {
				Expect(os.Setenv("BPL_DEBUG_PORT", "8001")).To(Succeed())
			})

			it.After(func() {
				Expect(os.Unsetenv("BPL_DEBUG_PORT")).To(Succeed())
			})

			it("contributes port configuration from $BPL_DEBUG_PORT", func() {
				Expect(d.Execute()).To(Equal(map[string]string{
					"JAVA_TOOL_OPTIONS": "-agentlib:jdwp=transport=dt_socket,server=y,address=*:8001,suspend=n",
				}))
			})
		})

		context("$BPL_DEBUG_SUSPEND", func() {
			it.Before(func() {
				Expect(os.Setenv("BPL_DEBUG_SUSPEND", "true")).To(Succeed())
			})

			it.After(func() {
				Expect(os.Unsetenv("BPL_DEBUG_SUSPEND")).To(Succeed())
			})

			it("contributes suspend configuration from $BPL_DEBUG_SUSPEND", func() {
				Expect(d.Execute()).To(Equal(map[string]string{
					"JAVA_TOOL_OPTIONS": "-agentlib:jdwp=transport=dt_socket,server=y,address=*:8000,suspend=y",
				}))
			})
		})

		context("$JAVA_TOOL_OPTIONS", func() {
			it.Before(func() {
				Expect(os.Setenv("JAVA_TOOL_OPTIONS", "test-java-tool-options")).To(Succeed())
			})

			it.After(func() {
				Expect(os.Unsetenv("JAVA_TOOL_OPTIONS")).To(Succeed())
			})

			it("contributes configuration appended to existing $JAVA_TOOL_OPTIONS", func() {
				Expect(d.Execute()).To(Equal(map[string]string{
					"JAVA_TOOL_OPTIONS": "test-java-tool-options -agentlib:jdwp=transport=dt_socket,server=y,address=*:8000,suspend=n",
				}))
			})
		})

		context("IPv6 is not present", func() {
			it.Before(func() {
				d1 := []byte("1")
				os.WriteFile(fakeIPv6File.Name(), d1, 0644)
			})

			it.After(func() {
				d1 := []byte("0")
				os.WriteFile(fakeIPv6File.Name(), d1, 0644)
			})

			it("replaces '*' host with IPv4 0.0.0.0", func() {
				Expect(d.Execute()).To(Equal(map[string]string{
					"JAVA_TOOL_OPTIONS": "-agentlib:jdwp=transport=dt_socket,server=y,address=0.0.0.0:8000,suspend=n",
				}))
			})
		})

		context("IPv6 kernel module file not there", func() {
			it.Before(func() {
				d.CustomIPv6CheckPath = "/does/not/exist"
			})

			it.After(func() {
				d.CustomIPv6CheckPath = fakeIPv6File.Name()
			})

			it("replaces '*' host with IPv4 0.0.0.0", func() {
				Expect(d.Execute()).To(Equal(map[string]string{
					"JAVA_TOOL_OPTIONS": "-agentlib:jdwp=transport=dt_socket,server=y,address=0.0.0.0:8000,suspend=n",
				}))
			})
		})

	})
}
