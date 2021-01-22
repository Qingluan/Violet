[toc]

# Violet Browser controller


## Language


### get
    args1: url


### click
    (if in each loop , do not need args1)
    args1: cssselector/ xpath selector / text

> example

```py
click: "登录/注册"
```

### back
    back to last page

> example

```py
back
```


### input

    (if in each loop , do not need args1)
    args1: cssselector/ xpath selector / text
    args2: text
    
    kargs:
        (if set , args1 , args2 will not work)
        name: None
        password: None
        end: "\n" # will to find button click

> example

```py
# normal 
input: input#name , "name"
input: //input[@type="password"] , "pwd"

# smart auth

input: name="name", password= "password", end="\n" # if add end ,will find btn to submit

```


### scroll
    [args1: cssselector/ xpath selector / text]


> example

```py
# scroll to bottom 
scroll

# scoll to some ele
scroll : 下一页
```

### savescreen
    args1: file path

### load
    kargs:
        url: url string # if set will get to this url
        cookie: cookie header string

### wait

    args1: [cssselector/ xpath selector / text]

    kargs:  
        sleep = 10
        change = url # will wait util url change

### for  ...[code]... end

    args1: [cssselector/ xpath selector / text]
        if args1 exists : will loop


### each  ...[ele code]... end

    args1: [cssselector/ xpath selector / text]
    
```py
for ele in smart_find_eles(args1):
    ...[ele code]...

```

#### ele code:
    kargs: 
        find : sub css/xpath/ text
        attrs : attrs in node
        contains: str
```
            # if str in this.Text():
            #    go on
            # else:
            #    break
```

##### save
    args1: file path

    kargs:
        as "ele code"