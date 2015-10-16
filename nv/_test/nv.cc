#include <array>
#include <iomanip>
#include <map>
#include <iostream>
#include <sstream>

#include <math.h>
#include <assert.h>
#include <stdint.h>
#include <stdio.h>
#include <string.h>

#include <libnvpair.h>

#define stringify(s) xstringify(s)
#define xstringify(s) #s

#define fnvlist_add_double(l, n, v) assert(nvlist_add_double(l, n, v) == 0)
#define fnvlist_add_hrtime(l, n, v) assert(nvlist_add_hrtime(l, n, v) == 0)

std::stringstream tests;

static void print(nvlist_t *list, char *name) {
	char *buf = NULL;
	size_t blen;
	int err;
	if ((err = nvlist_pack(list, &buf, &blen, NV_ENCODE_XDR, 0)) != 0) {
		std::cerr << "nvlist_pack error:" << err << "\n";
	}

	tests << "\t{name: \"" << name << "\", payload: []byte(\"";

	for (unsigned i = 0; i < blen; i++) {
		tests << "\\x"
			<< std::hex << std::setw(2) << std::setfill('0') << std::right
			<< (buf[i] & 0xFF);
	}
	tests << "\")},\n";
}

char *stra(char *s, int n) {
	size_t sl = strlen(s) + 2;
	char scomma[sl];
	memset(scomma, '\0', sl);
	strcat(scomma, s);
	strcat(scomma, ",");
	sl -= 1; // remove accounting for '\0'

	size_t size = sl * n + 1; // is really 1 byte extra but makes later stuff easier
	s = (char *)calloc(1, size);
	assert(s);

	for (int i = 0; i < n; i++) {
		strcat(s, scomma);
	}
	s[size - 2] = '\0'; //overwrite trailing ',' with '\0'
	return s;
}

char *stru(unsigned long long i) {
	char *s = NULL;
	int err = asprintf(&s, "%llu", i);
	if (err == -1) {
		std::cerr << "asprintf error:" << err << "\n";
		assert(err != -1);
	}
	return s;
}

char *stri(long long i) {
	char *s = NULL;
	int err = asprintf(&s, "%lld", i);
	if (err == -1) {
		std::cerr << "asprintf error:" << err << "\n";
		assert(err != -1);
	}
	return s;
}

char *strf(double d) {
	char *s = NULL;
	int err = asprintf(&s, "%16.17g", d);
	if (err == -1) {
		std::cerr << "asprintf error:" << err << "\n";
		assert(err != -1);
	}
	return s;
}

#define do_signed(lower, UPPER) do { \
	std::map<std::string, lower##_t> map; \
	map[stri(UPPER##_MIN)] = UPPER##_MIN; \
	map[stri(UPPER##_MIN+1)] = UPPER##_MIN+1; \
	map[stri(UPPER##_MIN/2)] = UPPER##_MIN/2; \
	map["-1"] = -1; \
	map["0"] = 0; \
	map["1"] = 1; \
	map[stri(UPPER##_MAX/2)] = UPPER##_MAX/2; \
	map[stri(UPPER##_MAX-1)] = UPPER##_MAX-1; \
	map[stri(UPPER##_MAX)] = UPPER##_MAX; \
	l = fnvlist_alloc(); \
	for (auto &kv : map) { \
		fnvlist_add_##lower(l, kv.first.c_str(), kv.second); \
	} \
	print(l, stringify(lower) "s"); \
	fnvlist_free(l); \
} while(0)

#define do_unsigned(lower, UPPER) do { \
	std::map<std::string, lower##_t> map; \
	map["0"] = 0; \
	map["1"] = 1; \
	map[stru(UPPER##_MAX/2)] = UPPER##_MAX/2; \
	map[stru(UPPER##_MAX-1)] = UPPER##_MAX-1; \
	map[stru(UPPER##_MAX)] = UPPER##_MAX; \
	l = fnvlist_alloc(); \
	for (auto &kv : map) { \
		fnvlist_add_##lower(l, kv.first.c_str(), kv.second); \
	} \
	print(l, stringify(lower) "s"); \
	fnvlist_free(l); \
} while(0)

#define do_double(lower, UPPER) do { \
	std::map<std::string, lower> map; \
	double offset = .9999;\
	map[strf(UPPER##_MIN)] = UPPER##_MIN; \
	map[strf(UPPER##_MIN/offset)] = UPPER##_MIN/offset; \
	map[strf(UPPER##_MIN/2)] = UPPER##_MIN/2; \
	map["-1"] = -1; \
	map["0"] = 0; \
	map["1"] = 1; \
	map[strf(UPPER##_MAX/2)] = UPPER##_MAX/2; \
	map[strf(UPPER##_MAX*offset)] = UPPER##_MAX*offset; \
	map[strf(UPPER##_MAX)] = UPPER##_MAX; \
	map[strf(M_E)] = M_E; \
	map[strf(M_LOG2E)] = M_LOG2E; \
	map[strf(M_LOG10E)] = M_LOG10E; \
	map[strf(M_LN2)] = M_LN2; \
	map[strf(M_LN10)] = M_LN10; \
	map[strf(M_PI)] = M_PI; \
	map[strf(M_PI_2)] = M_PI_2; \
	map[strf(M_PI_4)] = M_PI_4; \
	map[strf(M_1_PI)] = M_1_PI; \
	map[strf(M_2_SQRTPI)] = M_2_SQRTPI; \
	map[strf(M_SQRT2)] = M_SQRT2; \
	map[strf(M_SQRT1_2)] = M_SQRT1_2; \
	l = fnvlist_alloc(); \
	for (auto &kv : map) { \
		fnvlist_add_##lower(l, kv.first.c_str(), kv.second); \
	} \
	print(l, stringify(lower) "s"); \
	fnvlist_free(l); \
} while(0)

#define arrset(array, alen, val) do { \
	size_t i = 0; \
	for (i = 0; i < alen; i++) { \
		array[i] = val; \
	} \
} while (0)

#define do_signed_array(len, lower, UPPER) do { \
	typedef std::array<lower##_t, len> array; \
	array arr; \
	std::map<std::string, array> map; \
	arr.fill(UPPER##_MIN); map[stra(stri(UPPER##_MIN), arr.size())] = arr; \
	arr.fill(UPPER##_MIN+1); map[stra(stri(UPPER##_MIN+1), arr.size())] = arr; \
	arr.fill(UPPER##_MIN/2); map[stra(stri(UPPER##_MIN/2), arr.size())] = arr; \
	arr.fill(-1), map[stra(stri(-1), len)] = arr; \
	arr.fill(0); map[stra(stri(0), arr.size())] = arr; \
	arr.fill(1); map[stra(stri(1), arr.size())] = arr; \
	arr.fill(UPPER##_MAX/2); map[stra(stri(UPPER##_MAX/2), arr.size())] = arr; \
	arr.fill(UPPER##_MAX-1); map[stra(stri(UPPER##_MAX-1), arr.size())] = arr; \
	arr.fill(UPPER##_MAX); map[stra(stri(UPPER##_MAX), arr.size())] = arr; \
	l = fnvlist_alloc(); \
	for (auto &kv : map) { \
		fnvlist_add_##lower##_array(l, kv.first.c_str(), kv.second.data(), kv.second.size()); \
	} \
	print(l, stringify(lower) " array(" stringify(len)")"); \
	fnvlist_free(l); \
} while(0) \

#define do_unsigned_array(len, lower, UPPER) do { \
	typedef std::array<lower##_t, len> array; \
	array arr; \
	std::map<std::string, array> map; \
	arr.fill(0); map[stra(stru(0), arr.size())] = arr; \
	arr.fill(1); map[stra(stru(1), arr.size())] = arr; \
	arr.fill(UPPER##_MAX/2); map[stra(stru(UPPER##_MAX/2), arr.size())] = arr; \
	arr.fill(UPPER##_MAX-1); map[stra(stru(UPPER##_MAX-1), arr.size())] = arr; \
	arr.fill(UPPER##_MAX); map[stra(stru(UPPER##_MAX), arr.size())] = arr; \
	l = fnvlist_alloc(); \
	for (auto &kv : map) { \
		fnvlist_add_##lower##_array(l, kv.first.c_str(), kv.second.data(), kv.second.size()); \
	} \
	print(l, stringify(lower)" array"); \
	fnvlist_free(l); \
} while(0) \

int main() {
	tests <<"package nv\n"
		"\n"
		"/* !!! GENERATED FILE DO NOT EDIT !!! */\n"
		"\n"
		"var good = []struct {\n"
		"\tname    string\n"
		"\tpayload []byte\n"
		"}{\n";

	nvlist_t *l = NULL;
	{
		l = fnvlist_alloc();
		print(l, "empty");
		fnvlist_free(l);
	}
	{
		l = fnvlist_alloc();
		fnvlist_add_boolean_value(l, "false", B_FALSE);
		fnvlist_add_boolean_value(l, "true", B_TRUE);
		print(l, "bools");
		fnvlist_free(l);
	}
	{
		l = fnvlist_alloc();
		size_t len = 5;
		boolean_t array[len];
		arrset(array, len, B_FALSE); fnvlist_add_boolean_array(l, "false,false,false,false,false", array, len);
		arrset(array, len, B_TRUE); fnvlist_add_boolean_array(l, "true,true,true,true,true", array, len);
		print(l, "bool array");
		fnvlist_free(l);
	}

	l = fnvlist_alloc();
	fnvlist_add_byte(l, "-128", -128);
	fnvlist_add_byte(l, "0", 0);
	fnvlist_add_byte(l, "1", 1);
	fnvlist_add_byte(l, "127", 127);
	print(l, "bytes");
	fnvlist_free(l);

	{
		l = fnvlist_alloc();
		size_t len = 5;
		unsigned char array[len];
		arrset(array, len, -128); fnvlist_add_byte_array(l, "-128,-128,-128,-128,-128", array, len);
		arrset(array, len, 0); fnvlist_add_byte_array(l, "0,0,0,0,0", array, len);
		arrset(array, len, 1); fnvlist_add_byte_array(l, "1,1,1,1,1", array, len);
		arrset(array, len, 127); fnvlist_add_byte_array(l, "127,127,127,127,127", array, len);
		print(l, "byte array");
		fnvlist_free(l);
	}


	do_signed(int8, INT8);
	do_signed(int16, INT16);
	do_signed(int32, INT32);
	do_signed(int64, INT64);
	do_signed_array(4, int8, INT8);
	do_signed_array(4, int16, INT16);
	do_signed_array(4, int32, INT32);
	do_signed_array(4, int64, INT64);
	do_signed_array(5, int8, INT8);
	do_signed_array(5, int16, INT16);
	do_signed_array(5, int32, INT32);
	do_signed_array(5, int64, INT64);

	do_unsigned(uint8, UINT8);
	do_unsigned(uint16, UINT16);
	do_unsigned(uint32, UINT32);
	do_unsigned(uint64, UINT64);
	do_unsigned_array(4, uint8, UINT8);
	do_unsigned_array(4, uint16, UINT16);
	do_unsigned_array(4, uint32, UINT32);
	do_unsigned_array(4, uint64, UINT64);
	do_unsigned_array(5, uint8, UINT8);
	do_unsigned_array(5, uint16, UINT16);
	do_unsigned_array(5, uint32, UINT32);
	do_unsigned_array(5, uint64, UINT64);

	l = fnvlist_alloc();
	fnvlist_add_string(l, "0", "0");
	fnvlist_add_string(l, "1", "1");
	fnvlist_add_string(l, "HiMom", "HiMom");
	fnvlist_add_string(l, "\xff\"; DROP TABLE USERS;", "\xff\"; DROP TABLE USERS;");
	print(l, "strings");
	fnvlist_free(l);


	{
		char *array[] = {
			"0",
			"1",
			"HiMom",
			"\xff\"; DROP TABLE USERS;",
		};
		l = fnvlist_alloc();
		fnvlist_add_string_array(l, "0,1,HiMom,\xff\"; DROP TABLE USERS;", array, 4);
		print(l, "string array");
		fnvlist_free(l);
	}

	do_signed(hrtime, INT64);

	l = fnvlist_alloc();
	nvlist_t *le = fnvlist_alloc();
	fnvlist_add_boolean_value(le, "false", B_FALSE);
	fnvlist_add_boolean_value(le, "true", B_TRUE);
	fnvlist_add_nvlist(l, "nvlist", le);
	print(l, "nvlist");
	fnvlist_free(l);

	/*
	{
		l = fnvlist_alloc();
		nvlist_t *larr[] = {};
		fnvlist_add_nvlist_array(l, "empty", larr, 0);
		print(l, "empty nvlist array");
		fnvlist_free(l);
	}
	*/
	{
		l = fnvlist_alloc();
		nvlist_t *larr[] = {le, le};
		fnvlist_add_nvlist_array(l, "list,list", larr, 2);
		print(l, "nvlist array");
		fnvlist_free(le);
		fnvlist_free(l);
	}

	do_double(double, DBL);

	tests << "}\n";

	std::cout << tests.str() << std::flush;
	return 0;
}
