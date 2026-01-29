CC=gcc
CFLAGS=-fPIC -shared -ldl
TARGET_DIR=bin
TARGET=$(TARGET_DIR)/amux-hook.so

all: $(TARGET)

$(TARGET): hook.c
	mkdir -p $(TARGET_DIR)
	$(CC) $(CFLAGS) -o $@ $<

clean:
	rm -rf $(TARGET_DIR)
