package Funcs

var (
	Docs = map[string]interface{}{
		"get": `
	*@url 
	@@timeout:int
		`,
		"wait": `
	*@selector
	@@timeout:int
	@@change=
	`,
		"scroll": `
		if null will scroll download to bottom
	@selector
	@@offset=int`,
		"save": `
	@filename
	@@cookie=
	`,
		"if": `
	*@selector
	@attr:str
	@compareStr:str
	test @selector.attr == comparestr `,
		"for": `
	@existed_selector_or_loop_int
		if selecotr exists or in loop limit .
	`,
		"input": `
	*@selector: str
	@inputcontent: str   if set @@name ... this is will not work
	@@name:str @@password:str  auto detect username/password input to type.
	@@end:str if set will submit
	`,
		"each": map[string]string{
			"save": `
			save node's data to file, in json format
		@filename
		@@contains:str
		@@find=subselector:str
		@@attrs=[attrs...]`,
		},
	}
)
