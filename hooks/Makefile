CC=gcc
CFLAGS=-fPIC -shared -ldl -D_GNU_SOURCE
TARGET=bin/amux-hook.so
SRC=hook.c

all: $(TARGET)

$(TARGET): $(SRC)
	mkdir -p bin
	$(CC) $(CFLAGS) -o $@ $<

clean:
	rm -f $(TARGET)