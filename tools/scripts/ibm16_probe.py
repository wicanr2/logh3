# 影格的編碼是哪一種 RLE？判準：同一個變體套到 374 格，解出來的長度
# 應該收斂到少數幾個定值（一個 plane 18160、三個 54480…）。
import collections, struct
lens=list(struct.unpack_from('<374H', open('/orig/IBMFLS16.350','rb').read(), 0))
d=open('/orig/IBM16.350','rb').read()
offs=[]; a=0
for v in lens: offs.append(a); a+=v

def run_len(buf, mode):
    p=0; out=0; prev=-1; n=len(buf)
    while p<n:
        v=buf[p]; p+=1; out+=1; prev=v
        while p<n:
            v2=buf[p]; p+=1
            if v2!=prev:
                out+=1; prev=v2; continue
            if p>=n: break
            if mode=='word':
                if p+2>n: break
                c=struct.unpack_from('<H',buf,p)[0]; p+=2
            else:
                c=buf[p]; p+=1
                if mode=='esc' and c==0:
                    if p+2>n: break
                    c=struct.unpack_from('<H',buf,p)[0]; p+=2
            out+=c-1
            break
    return out

for mode in ('esc','byte','word'):
    got=collections.Counter()
    for i in range(374):
        got[run_len(d[offs[i]:offs[i]+lens[i]], mode)]+=1
    print('%-5s 前 6 名長度：%s'%(mode, got.most_common(6)))
