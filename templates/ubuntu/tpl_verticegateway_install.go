/*
** Copyright [2013-2016] [Megam Systems]
**
** Licensed under the Apache License, Version 2.0 (the "License");
** you may not use this file except in compliance with the License.
** You may obtain a copy of the License at
**
** http://www.apache.org/licenses/LICENSE-2.0
**
** Unless required by applicable law or agreed to in writing, software
** distributed under the License is distributed on an "AS IS" BASIS,
** WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
** See the License for the specific language governing permissions and
** limitations under the License.
 */

package ubuntu

import (
	"github.com/megamsys/libmegdc/templates"
	"github.com/megamsys/urknall"
)

var ubuntugatewayinstall *UbuntuGatewayInstall

func init() {
	ubuntugatewayinstall = &UbuntuGatewayInstall{}
	templates.Register("UbuntuGatewayInstall", ubuntugatewayinstall)
}

type UbuntuGatewayInstall struct{
			hostip string
}

func (tpl *UbuntuGatewayInstall) Render(p urknall.Package) {
	p.AddTemplate("verticegateway", &UbuntuGatewayInstallTemplate{
				hostip: tpl.hostip,
	})
}

func (tpl *UbuntuGatewayInstall) Options(t *templates.Template) {
	if host,ok := t.Options[HOST]; ok {
		tpl.hostip = host
	}
}

func (tpl *UbuntuGatewayInstall) Run(target urknall.Target,inputs map[string]string) error {
	return urknall.Run(target, &UbuntuGatewayInstall{},inputs)
}

type UbuntuGatewayInstallTemplate struct{
			hostip string
}

func (m *UbuntuGatewayInstallTemplate) Render(pkg urknall.Package) {

	pkg.AddCommands("install",
		InstallPackages("verticegateway"),
	)
}
