import sys

from multiversx_sdk import Address


def main():
    # input arguments
    address = Address.new_from_hex(sys.argv[1])
    print(address.to_bech32())


if __name__ == "__main__":
    main()
