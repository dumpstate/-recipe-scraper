.PHONY: clean build

all: build

RM	=	rm -rf
MKD	=	mkdir -p
GO	=	go
TAR	=	tar

NAME	=	recipe-scraper
BIN		=	./bin

GO_SRC	=	$(shell find ./ -type f -name '*.go')

clean:
	$(RM) $(BIN)

build: $(BIN) $(BIN)/$(NAME)

package: build
	$(TAR) cfJ $(NAME).tar.xz $(BIN)/$(NAME) LICENSE README.md

$(BIN):
	$(MKD) $(BIN)

$(BIN)/$(NAME): $(GO_SRC)
	$(GO) build -o $(BIN)/$(NAME) $(GO_SRC)
