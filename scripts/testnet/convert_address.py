import sys

from multiversx_sdk import Address


def main():
    # input arguments
    address = Address.from_bech32(sys.argv[1])
    new_hrp = sys.argv[2]

    print(Address(address.pubkey, new_hrp).to_bech32())


if __name__ == "__main__":
    main()
