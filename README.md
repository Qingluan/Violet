[toc]

# Violet Browser controller


## Language


### get
    args1: url


### click
    (if in each loop , do not need args1)
    args1: cssselector/ xpath selector / text

### input

    (if in each loop , do not need args1)
    args1: cssselector/ xpath selector / text
    args2: text
    
    kargs:
        (if set , args1 , args2 will not work)
        name: None
        password: None
        end: None

### back
    back to last

### scroll
    [args1: cssselector/ xpath selector / text]

### savescreen
    args1: file path


### wait

    args1: [cssselector/ xpath selector / text]

    kargs:  sleep = 4

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
    

##### save
    args1: file path

    kargs:
        as "ele code"