# Captcha

an example for use captcha

```
package controllers

import (
	"github.com/achatur/beego"
	"github.com/achatur/beego/cache"
	"github.com/achatur/beego/utils/captcha"
)

var cpt *captcha.Captcha

func init() {
	// use beego cache system store the captcha data
	store := cache.NewMemoryCache()
	cpt = captcha.NewWithFilter("/captcha/", store)
}

type MainController struct {
	beego.Controller
}

func (this *MainController) Get() {
	this.TplNames = "index.tpl"
}

func (this *MainController) Post() {
	this.TplNames = "index.tpl"

	this.Data["Success"] = cpt.VerifyReq(this.Ctx.Request)
}
```

template usage

```
{{.Success}}
<form action="/" method="post">
	{{create_captcha}}
	<input name="captcha" type="text">
</form>
```
