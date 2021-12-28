# lookup-addrs

Performs quick intel of IP addresses supplied.
If possible, grabs: 
- subject name of SSL certificate
- organization name of SSL certificate
- ip addr lookup result (equivalent to nslookup result), which is cloud hosting provider

Works quite similarly to `shodan parse`, but much simpler version.

## Usage
- Input a file with IPs in it. Doesn't matter if it's in the form of a complicated URL like `https://3.3.3.3:12456/asldfkjdklsajweaf?125125=1`.
- It will just simply pull out IPs by using regex (fine print: therefore, it will basically pull out anything that looks like an IP, even if it is in the querystring of the URL, for example)

# Build
```
go build
```

# Run
```
Usage of lookup-addrs:
  -inputFilePath string
        Path to the text file that contains list of URLs (default "./urls.txt")
  -outputFilePath string
        Path to output CSV file (default "./out.csv")
  -threads int
        Number of threads to use (default 20)
  -verbose
        Verbosity (default true)
Example: 
./lookup-addrs -threads=30 -inputFilePath=input.txt -outputFilePath=output.csv -verbose=false

Note that the boolean flag should be fed as -isVerbose=false. -isVerbose false won't get it to work.
```

# Example input & output
Input:

```
100.24.148.45:443
100.24.154.46:80
100.24.158.213:80
100.24.169.157:80
100.24.176.122:80
100.24.176.59:80
100.24.197.246:80
100.24.202.245:80
100.24.214.180:80
100.24.217.207:443
100.24.222.19:443
100.24.230.119:80
100.24.241.246:80
100.24.251.5:80
100.24.77.124:80
100.24.78.201:443
100.24.83.233:80
100.24.85.28:80
100.24.93.146:80
100.25.106.49:443
100.25.176.220:80
100.25.185.7:443
100.25.186.116:443
100.25.186.116:80
100.25.194.119:443
100.25.196.233:80
100.25.197.121
```

```
100.24.148.45:443,,data.infopriceti.com.br,
100.24.154.46:80,,,
100.24.158.213:80,,,
100.24.169.157:80,,,
100.24.176.122:80,,,
100.24.176.59:80,,,
100.24.197.246:80,,,
100.24.202.245:80,,,
100.24.214.180:80,,,
100.24.217.207:443,,sbopslive.servisbotconnect.com,
100.24.222.19:443,,operations.vidgo.veygo.co,
100.24.230.119:80,,,
100.24.241.246:80,,,
100.24.251.5:80,,,
100.24.77.124:80,,,
100.24.78.201:443,,fivana.com,
100.24.83.233:80,,,
100.24.85.28:80,,,
100.24.93.146:80,,,
100.25.106.49:443,,oncoramedical.com,
100.25.176.220:80,,,
100.25.185.7:443,,accessclinicaltrials.niaid.nih.gov,National Institutes of Health,Entrust, Inc.,Entrust, Inc.
100.25.186.116:443,,gct-internal.net,
100.25.186.116:80,,,
100.25.194.119:443,,grafana.prudentte.com.br,
100.25.196.233:80,,,
100.25.197.121,,,
```