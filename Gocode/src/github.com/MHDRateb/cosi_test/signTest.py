from ctypes import *
import ctypes
lib = cdll.LoadLibrary("./cosiTest.so")

# define class GoString to map:
# C type struct { const char *p; GoInt n; }
class GoString(Structure):
    _fields_ = [("p", c_char_p), ("n", c_longlong)]


    
lib.startSign.argtypes = [GoString]
lib.startSign.restype = c_longlong

message =b"Hello Python!"

msg = GoString( message, len(message))
signaggr = lib.startSign(msg)
print ('returned value',ctypes.string_at(signaggr))




