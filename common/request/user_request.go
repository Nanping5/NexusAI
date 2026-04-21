package request

type UserRegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Captcha  string `json:"captcha" binding:"required,len=6"`
}

type UserLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type SendCaptchaRequest struct {
	Email string `json:"email" binding:"required,email"`
}
