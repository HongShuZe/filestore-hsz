package handler


import (
	"net/http"
	"filestore-hsz/util"

	dblayer "filestore-hsz/db"
	"fmt"
	"time"
)

const (
	// 用于加密的盐值
	pwdSalt = "#890"
)

// 处理用户注册请求
func SignupHandler(w http.ResponseWriter, r *http.Request)  {
	// GET请求重定向到注册页面
	if r.Method == http.MethodGet {
		http.Redirect(w, r, "/static/view/signup.html", http.StatusFound)
		return
	}

	r.ParseForm()
	username := r.Form.Get("username")
	passwd := r.Form.Get("password")

	if len(username) < 3 || len(passwd) < 5 {
		w.Write([]byte("Invalid parameter"))
		return
	}

	// 对密码进行加盐及sha1值加密
	encPasswd := util.Sha1([]byte(passwd + pwdSalt))
	// 将用户信息注册到用户表中
	suc := dblayer.UserSignup(username, encPasswd)
	if suc {
		w.Write([]byte("SUCCESS"))
	}else {
		w.Write([]byte("FAILED"))
	}

}

// 登录接口
func SignInHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.Redirect(w, r, "/static/view/signin.html", http.StatusFound)
		return
	}

	r.ParseForm()
	username := r.Form.Get("username")
	password := r.Form.Get("password")
	encPasswd := util.Sha1([]byte(password + pwdSalt))

	// 1.校验用户名及密码
	pwdChecked := dblayer.UserSignin(username, encPasswd)
	if !pwdChecked {
		w.Write([]byte("FAILED"))
		return
	}
	// 2.生成访问凭证
	token := GenToken(username)
	upRes := dblayer.UpdateToken(username, token)
	if !upRes {
		w.Write([]byte("FAILED"))
		return
	}

	// 3.登录成功重定向到首页
	// w.Write([]byte("http://" + r.Host + "/static/view/home.html"))
	resp := util.RespMsg{
		Code: 0,
		Msg: "OK",
		Data: struct {
			Location string
			Username string
			Token    string
		}{
			Location: "http://" + r.Host + "/static/view/home.html",
			Username: username,
			Token: token,
		},
	}

	w.Write(resp.JSONBytes())
}

// 查询用户信息
func UserInfoHandler(w http.ResponseWriter, r *http.Request)  {
	// 1.解析请求参数
	r.ParseForm()
	username := r.Form.Get("username")
	//token := r.Form.Get("token")

	// 2.验证token是否有效(在拦截器中已经判断了)
	//isValidToken := IsTokenValid(token)
	//if !isValidToken {
	//	w.WriteHeader(http.StatusForbidden)
	//	return
	//}

	// 3.查询用户信息
	user, err := dblayer.GetUserInfo(username)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	// 4.组装并响应用户数据
	resp := util.RespMsg{
		Code: 0,
		Msg: "OK",
		Data: user,
	}
	w.Write(resp.JSONBytes())
}

// 生成token
func GenToken(username string) string {
	ts := fmt.Sprintf("%x", time.Now().Unix())
	tokenPrefix := util.MD5([]byte(username + ts + "_tokensalt"))
	return tokenPrefix + ts[:8]
}

// token是否有效
func IsTokenValid(token, username string) bool {
	if len(token) != 40 {
		return false
	}
	// TODO 判断token的时效性， 是否过期
	// 假设token的有效期为1天
	tokenTS := token[:8]
	if util.Hex2Dec(tokenTS) < time.Now().Unix()-86400 {
		return false
	}

	// TODO 从数据库表tbl_user_token查询username对应的token信息
	// TODO 对比两个token是否一致
	if dblayer.GetUserToken(username) != token {
		return false
	}

	return true
}