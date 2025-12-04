// revive:disable:var-naming
package utils

// revive:enable:var-naming

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func ExcludeEmptyValuesFromArray(arr pulumi.StringArray) pulumi.StringArrayOutput {
	return arr.ToStringArrayOutput().ApplyT(func(arr []string) []string {
		l := make([]string, 0)
		for _, v := range arr {
			if v == "" {
				continue
			}
			l = append(l, v)
		}
		return l
	}).(pulumi.StringArrayOutput)
}
