import idc
with open("/work/hello.txt", "w") as fh:
    fh.write("ARGV=%r\n" % (list(idc.ARGV),))
idc.qexit(0)
