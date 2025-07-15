package template

import (
	"bytes"
	"html/template"
	"net/http"

	"chchma.com/cloudlite-sync/internal/utils"
)

type TemplateEngine struct {
	templates *template.Template
}

func New() *TemplateEngine {
	// 定义模板函数
	funcMap := template.FuncMap{
		"formatTime": func(t interface{}) string {
			// 这里可以添加时间格式化函数
			return t.(string)
		},
		"formatFileSize": utils.FormatFileSize, // 直接用 utils 里的函数
	}

	return &TemplateEngine{
		templates: template.New("").Funcs(funcMap),
	}
}

// Render 渲染模板
func (te *TemplateEngine) Render(w http.ResponseWriter, templateName string, data interface{}) error {
	// 动态加载 layout.html 和页面模板
	tmpl, err := te.templates.Clone()
	if err != nil {
		return err
	}

	// 先加载 layout.html
	_, err = tmpl.ParseFiles("templates/layout.html")
	if err != nil {
		return err
	}

	// 再加载页面模板
	_, err = tmpl.ParseFiles("templates/" + templateName)
	if err != nil {
		return err
	}

	// 渲染 layout.html（它会自动引用页面模板的 content 块）
	return tmpl.ExecuteTemplate(w, "layout.html", data)
}

// RenderString 渲染模板到字符串
func (te *TemplateEngine) RenderString(templateName string, data interface{}) (string, error) {
	var buf bytes.Buffer
	err := te.templates.ExecuteTemplate(&buf, templateName, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// PageData 页面数据结构
type PageData struct {
	Title      string
	User       string
	Data       interface{}
	Error      string
	Success    string
	Pagination *PaginationData
}

type PaginationData struct {
	CurrentPage int
	TotalPages  int
	TotalItems  int
	PageSize    int
	HasNext     bool
	HasPrev     bool
	NextPage    int
	PrevPage    int
}

// NewPageData 创建新的页面数据
func NewPageData(title string, data interface{}) *PageData {
	return &PageData{
		Title: title,
		Data:  data,
	}
}

// SetError 设置错误信息
func (pd *PageData) SetError(err string) {
	pd.Error = err
}

// SetSuccess 设置成功信息
func (pd *PageData) SetSuccess(msg string) {
	pd.Success = msg
}

// SetUser 设置用户信息
func (pd *PageData) SetUser(user string) {
	pd.User = user
}

// SetPagination 设置分页信息
func (pd *PageData) SetPagination(currentPage, totalItems, pageSize int) {
	totalPages := (totalItems + pageSize - 1) / pageSize

	pd.Pagination = &PaginationData{
		CurrentPage: currentPage,
		TotalPages:  totalPages,
		TotalItems:  totalItems,
		PageSize:    pageSize,
		HasNext:     currentPage < totalPages,
		HasPrev:     currentPage > 1,
		NextPage:    currentPage + 1,
		PrevPage:    currentPage - 1,
	}
}
