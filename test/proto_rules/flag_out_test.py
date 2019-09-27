import unittest
import os.path


class FlagOutTest(unittest.TestCase):

    def test_can_import(self):
        from test.proto_rules.flag_test_pb2 import FlagTestMessage
