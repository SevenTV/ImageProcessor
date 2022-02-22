package avif

/*
From: Wan-Teh Chang <wtc@google.com>
Date: Wed, 9 Feb 2022 10:40:45 -0800
Subject: Comment on the Test() function in https://github.com/SevenTV/ImageProcessor/blob/master/src/containers/avif/image.go
To: troybensonsa@gmail.com

Hi Troy,

Thank you again for filing the bug report. libavif is now very stable.
We haven't received a bug report for a while. I apologize for taking
more than two months to respond to your bug. It's a combination of
time-off, holidays, and negligence.

I am not sure if the Test() function in
https://github.com/SevenTV/ImageProcessor/blob/master/src/containers/avif/image.go
needs to be robust. It has some issues.

1. This is arguably a theoretical issue -- the AVIF file format does
not require the "avif" or "avis" brand to be in the major_brand field,
which is the bytes at indexes 8-11. After the major_brand field, there
is a variable-length compatible_brands list. If the "avif" or "avis"
brand is in compatible_brands, the file is also an AVIF image.

However, because of the prevalence of simple checks like this, I
believe all AVIF encoders must put the "avif" or "avis" brand in
major_brand to avoid compatibility issues. So you should be able to
ignore this issue in practice.

2. The first four bytes are the size field. The current code assumes
the size is 0x28. This is likely to break in the future. I suggest not
checking data[3] and perhaps not checking data[2] also.

3. The current code only checks for the "avis" brand, which means
"AVIF image sequence", i.e., AVIF animation. Would you like to also
check for the "avif" branch, which is used for AVIF still images?

Wan-Teh
*/

func Test(data []byte) bool {
	if len(data) < 12 {
		return false
	}

	return data[0] == 0x00 &&
		data[1] == 0x00 &&
		data[4] == 'f' &&
		data[5] == 't' &&
		data[6] == 'y' &&
		data[7] == 'p' &&
		data[8] == 'a' &&
		data[9] == 'v' &&
		data[10] == 'i' &&
		(data[11] == 's' || data[11] == 'f')
}
