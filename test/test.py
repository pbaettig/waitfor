import os
import os.path
import socket
import subprocess
import sys
from collections import namedtuple
from datetime import datetime, timedelta
from http.server import (BaseHTTPRequestHandler, HTTPServer,
                         SimpleHTTPRequestHandler)
from random import randint
from tempfile import NamedTemporaryFile, gettempdir
from threading import Thread
from time import sleep, time
from traceback import print_exc

Result = namedtuple('Result', ['out', 'err', 'runtime', 'returncode'])


def run_http_server(port, resp_status=200, resp_msg="testtesttest"):
    class myhandler(BaseHTTPRequestHandler):
        def do_GET(self):
            self.send_response(resp_status)
            self.send_header('Content-Type', 'text/plain; charset=utf-8')
            self.end_headers()
            if resp_msg != "":
                self.wfile.write(resp_msg.encode())
            return
        def log_message(self, format, *args):
            return

    server_address = ('127.0.0.1', port)
    httpd = HTTPServer(server_address, myhandler)
    httpd.handle_request()


class TestFailedException(Exception):
    pass


def build():
    current_dir = os.path.dirname(os.path.abspath(__file__))
    build_dir = os.path.join(os.path.dirname(current_dir), 'build')
    print(build_dir)
    out = subprocess.check_output(['./build.sh'], cwd=build_dir)
    print(out)
    #subprocess.call('go build -o wfor ../cmd/wfor/*.go', shell=True)


class DelayedTempFile(object):
    def __init__(self):
        self.name = os.path.join(gettempdir(), 'wfor-test-{}'.format(time()))
        self._fd = None
    
    def create(self):
        self._fd = open(self.name, 'w')

    def delete(self):
        os.remove(self.name)


class WforTest(object):
    executable_under_test = './wfor'
    def __init__(self, name, want_max_runtime=None, want_min_runtime=None, want_returncode=None, want_out=None):
        self.name = name
        self.want_min_runtime = want_min_runtime
        self.want_max_runtime = want_max_runtime
        self.want_returncode = want_returncode
        self.want_out = want_out


    def run(self):
        raise NotImplementedError("Needs to be implemented in subclass")


    def __call__(self):
        self.run()

    def check(self):
        got = self.run()
        
        msg_base = "Test {} failed. ".format(self.name)

        if self.want_min_runtime and got.runtime < self.want_min_runtime:
            raise TestFailedException(msg_base+"Wanted min. command runtime of {} seconds, but got {} seconds.".format(self.want_min_runtime.total_seconds(), got.runtime.total_seconds()))
        if self.want_max_runtime and got.runtime > self.want_max_runtime:
            raise TestFailedException(msg_base+"Wanted max. command runtime of {} seconds, but got {} seconds.".format(self.want_max_runtime.total_seconds(), got.runtime.total_seconds()))
        if self.want_returncode and self.want_returncode != got.returncode:
            raise TestFailedException(msg_base+"Wanted command returncode {}, but got {}.".format(self.want_returncode, got.returncode))
        if self.want_out and self.want_out != got.out:
            raise TestFailedException(msg_base+'Wanted command output "{}", but got "{}".'.format(self.want_out, got.out))
    
    def _start_wfor(self,args):
        p = subprocess.Popen([WforTest.executable_under_test]+args, stderr=subprocess.PIPE, stdout=subprocess.PIPE)
        return datetime.now(),p

    def _wait_wfor(self,start,p):
        out, err = p.communicate()
        runtime = datetime.now() - start
        return Result(out.decode(), err.decode(), runtime, p.returncode)


class PathWaitTimeoutNoExec(WforTest):
    def __init__(self):
        super().__init__("PathWaitTimeoutNoExec", want_max_runtime=timedelta(seconds=5, milliseconds=100), want_returncode=1)

    def run(self):
        return self._wait_wfor(
            *self._start_wfor(
                ['-path', './bubububu','-interval', '1s', '-timeout', '5s']))


class TcpWaitTimeoutNoExec(WforTest):
    def __init__(self):
        super().__init__("TcpWaitTimeoutNoExec", want_max_runtime=timedelta(seconds=6, milliseconds=100), want_returncode=1)

    def run(self):
        return self._wait_wfor(
            *self._start_wfor(
                ['-tcp', 'localhost:27641','-interval', '2s', '-timeout', '6s']))


class HttpWaitTimeoutNoExec(WforTest):
    def __init__(self):
        super().__init__("HttpWaitTimeoutNoExec", want_max_runtime=timedelta(seconds=4, milliseconds=100), want_returncode=1)

    def run(self):
        return self._wait_wfor(
            *self._start_wfor(
                ['-http', 'localhost:27641','-interval', '2s', '-timeout', '4s']))


class PathWaitSuccessWhileWaitingExec(WforTest):
    def __init__(self, want_returncode=123, want_max_runtime=timedelta(seconds=5, milliseconds=100), want_out="asfkljasklj"):
        super().__init__(
            "PathWaitSuccessWhileWaitingExec",
            want_max_runtime=want_max_runtime,
            want_returncode=want_returncode,
            want_out=want_out)

    def run(self):
        f = DelayedTempFile()

        s,p = self._start_wfor(
            ['-path', f.name,'-interval', '1s', '-timeout', '10s','--',
             '/bin/sh', '-c', 'echo -n "{}";exit {}'.format(self.want_out, self.want_returncode)
            ]
        )
        sleep(self.want_max_runtime.seconds-0.5)
        f.create()
        try:
            r = self._wait_wfor(s,p)
        except:
            print_exc()
        else:
            return r
        finally:
            f.delete()
        

class PathWaitSuccessNoExec(WforTest):
    def __init__(self):
        super().__init__("PathWaitSuccessNoExec", want_returncode=0)

    def run(self):
        with NamedTemporaryFile() as fd:
            r = self._wait_wfor(*self._start_wfor(['-path', fd.name, '-interval', '1s', '-timeout', '5s']))
        return r


class PathWaitSuccessExec(WforTest):
    def __init__(self, want_returncode=45, want_out="asfkljasklj"):
        super().__init__(
            "PathWaitSuccessExec",
            want_returncode=want_returncode,
            want_out=want_out)


    def run(self):
        with NamedTemporaryFile() as fd:
            r = self._wait_wfor(*self._start_wfor(
                ['-path', fd.name, '-interval', '1s', '-timeout', '5s',
                '/bin/sh', '-c', 'echo -n "{}";exit {}'.format(self.want_out, self.want_returncode)]))
        return r


class TcpWaitSuccessWhileWaitingExec(WforTest):
    def __init__(self, want_returncode=12, want_out="klskcmkshq"):
        super().__init__(
            "TcpWaitSuccessWhileWaitingExec",
            want_max_runtime=timedelta(seconds=3, milliseconds=100),
            want_returncode=want_returncode,
            want_out=want_out)

    def run(self):
        host = '127.0.0.1'
        port = randint(32000, 65000)
        s,p = self._start_wfor(
            ['-tcp', '{}:{}'.format(host,port),'-interval', '1s', '-timeout', '5s','--',
             '/bin/sh', '-c', 'echo -n "{}";exit {}'.format(self.want_out, self.want_returncode)
            ]
        )
        sleep(self.want_max_runtime.seconds - 0.5)
        
        with socket.socket(socket.AF_INET,socket.SOCK_STREAM) as sock:
            sock.bind((host, port))
            sock.listen()
            conn, _ = sock.accept()
            conn.close()
        
        try:
            r = self._wait_wfor(s,p)
        except:
            print_exc()
        else:
            return r


class HttpWaitSuccessExec(WforTest):
    def __init__(self, want_returncode=12, want_out="sdfsqen"):
        super().__init__(
            "HttpWaitSuccessExec",
            want_returncode=want_returncode,
            want_out=want_out)
    def run(self):
        port = randint(32000,65000)
        status = 200
        t = Thread(target=run_http_server, args=(port,), kwargs={'resp_status': status})
        t.start()
        r = self._wait_wfor(*self._start_wfor(
                ['-http', '127.0.0.1:{}|200-208'.format(port), '-interval', '1s', '-timeout', '5s',
                '/bin/sh', '-c', 'echo -n "{}";exit {}'.format(self.want_out, self.want_returncode)]))
        return r


class HttpWaitStatusSuccessWhileWaitingExec(WforTest):
    def __init__(self, want_returncode=71, want_max_runtime=timedelta(seconds=3, milliseconds=100), want_out="ksoqmm  spo"):
        super().__init__(
            "HttpWaitStatusSuccessWhileWaitingExec",
            want_max_runtime=want_max_runtime,
            want_returncode=want_returncode,
            want_out=want_out)

    def run(self):
        port = randint(32000,65000)
        status = 203
        http_server = Thread(target=run_http_server, args=(port,), kwargs={'resp_status': status})
        s,p = self._start_wfor(
            ['-http', '127.0.0.1:{}|200-208'.format(port),'-interval', '1s', '-timeout', '10s','--',
             '/bin/sh', '-c', 'echo -n "{}";exit {}'.format(self.want_out, self.want_returncode)
            ]
        )
        sleep(self.want_max_runtime.seconds-0.5)
        http_server.start()
        try:
            r = self._wait_wfor(s,p)
        except:
            print_exc()
        else:
            return r


class HttpWaitContentSuccessWhileWaitingExec(WforTest):
    def __init__(self, want_returncode=71, want_max_runtime=timedelta(seconds=4, milliseconds=100), want_out="ksoqmm  spo"):
        super().__init__(
            "HttpWaitContentSuccessWhileWaitingExec",
            want_max_runtime=want_max_runtime,
            want_returncode=want_returncode,
            want_out=want_out)

    def run(self):
        port = randint(32000,65000)

        http_server = Thread(target=run_http_server, args=(port,), kwargs={'resp_status': 200, 'resp_msg':'Test test test'})
        s,p = self._start_wfor(
            ['-http', '127.0.0.1:{}||Test .*'.format(port),'-interval', '1s', '-timeout', '10s','--',
             '/bin/sh', '-c', 'echo -n "{}";exit {}'.format(self.want_out, self.want_returncode)
            ]
        )
        sleep(self.want_max_runtime.seconds-0.5)
        http_server.start()
        try:
            r = self._wait_wfor(s,p)
        except:
            print_exc()
        else:
            return r


if __name__ == '__main__':
    if len(sys.argv) < 2:
        print("Building wfor...")
        build()
        exit(0)
    else:
        print("Running tests against {}".format(sys.argv[1]))
        WforTest.executable_under_test = sys.argv[1]

    tests = [
        HttpWaitStatusSuccessWhileWaitingExec(),
        HttpWaitContentSuccessWhileWaitingExec(),
        HttpWaitSuccessExec(),
        PathWaitSuccessNoExec(),
        PathWaitSuccessExec(),
        PathWaitTimeoutNoExec(),
        TcpWaitTimeoutNoExec(),
        HttpWaitTimeoutNoExec(),
        TcpWaitSuccessWhileWaitingExec(),
        PathWaitSuccessWhileWaitingExec()
    ]
    tests_failed = False
    for test in tests:
        print('Running {:<40s}'.format(test.name), end='', flush=True)
        try:
            test.check()
            print("Passed.")
        except TestFailedException as ex:
            tests_failed = True
            print("Failed. {}".format(ex))

    if tests_failed:
        exit(1)
    else:
        exit(0)
    #test_path_condition_fulfilled_during_wait()
    # test_timeout()
    # test_exec()
    # test_no_exec()
