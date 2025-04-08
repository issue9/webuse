package main

//go:generate web locale -l=und -m -f=yaml -o=./locales ./
//go:generate web update-locale -src=./locales/und.yaml -dest=./locales/zh.yaml

func main() {
	if err := Exec("id", "1.0.0");err!=nil{
		panic(err)
	}
}
