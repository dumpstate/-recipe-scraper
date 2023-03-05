.PHONY: clean build

all: build

RM	=	rm -rf
MKD	=	mkdir -p
GO	=	go

NAME	=	recipe-scraper
BIN		=	./bin

GO_SRC	=	$(shell find ./ -type f -name '*.go')

clean:
	$(RM) $(BIN)

build: $(BIN) $(BIN)/$(NAME)

$(BIN):
	$(MKD) $(BIN)

$(BIN)/$(NAME): $(GO_SRC)
	$(GO) build -o $(BIN)/$(NAME) $(GO_SRC)
