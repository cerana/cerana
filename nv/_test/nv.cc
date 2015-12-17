#include <array>
#include <iomanip>
#include <map>
#include <iostream>
#include <sstream>
#include <unordered_map>

#include <assert.h>
#include <ctype.h>
#include <math.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include <libnvpair.h>

#define stringify(s) xstringify(s)
#define xstringify(s) #s

#define fnvlist_add_double(l, n, v) assert(nvlist_add_double(l, n, v) == 0)
#define fnvlist_add_hrtime(l, n, v) assert(nvlist_add_hrtime(l, n, v) == 0)

std::stringstream defs;
std::stringstream tests;

std::unordered_map<int, const char *> types = {
	{DATA_TYPE_BOOLEAN, "Boolean"},
	{DATA_TYPE_BYTE, "byte"},
	{DATA_TYPE_INT16, "int16"},
	{DATA_TYPE_UINT16, "uint16"},
	{DATA_TYPE_INT32, "int32"},
	{DATA_TYPE_UINT32, "uint32"},
	{DATA_TYPE_INT64, "int64"},
	{DATA_TYPE_UINT64, "uint64"},
	{DATA_TYPE_STRING, "string"},
	{DATA_TYPE_BYTE_ARRAY, "[]byte"},
	{DATA_TYPE_INT16_ARRAY, "[]int16"},
	{DATA_TYPE_UINT16_ARRAY, "[]uint16"},
	{DATA_TYPE_INT32_ARRAY, "[]int32"},
	{DATA_TYPE_UINT32_ARRAY, "[]uint32"},
	{DATA_TYPE_INT64_ARRAY, "[]int64"},
	{DATA_TYPE_UINT64_ARRAY, "[]uint64"},
	{DATA_TYPE_STRING_ARRAY, "[]string"},
	{DATA_TYPE_HRTIME, "time.Duration"},
	{DATA_TYPE_NVLIST, "interface{}"},
	{DATA_TYPE_NVLIST_ARRAY, "[]map[string]interface{}"},
	{DATA_TYPE_BOOLEAN_VALUE, "bool"},
	{DATA_TYPE_INT8, "int8"},
	{DATA_TYPE_UINT8, "uint8"},
	{DATA_TYPE_BOOLEAN_ARRAY, "[]bool"},
	{DATA_TYPE_INT8_ARRAY, "[]int8"},
	{DATA_TYPE_UINT8_ARRAY, "[]uint8"},
	{DATA_TYPE_DOUBLE, "float64"},
};

// sanitize tags by escaping " and ` characters, and use escaped hex notation
// for non printable characters
static char *sanitize(char *name) {
	name = strdup(name);
	size_t slen = strlen(name);
	size_t buflen = strlen(name);
	// let the magic begin
	for (unsigned int i = 0; i < slen; i++) {
		assert(name[i] != '`');
		if (name[i] != '"' && isprint(name[i])) {
			continue;
		}
		if (slen + 2 >= buflen) {
			// grow the buffer
			buflen *= 2;
			char *newbuf = (char *)realloc(name, buflen);
			memset(newbuf+slen, 0, buflen-slen);
			name = newbuf;
		}
		if (name[i] == '"') {
			memmove(name+i+1, name+i, slen - i);
			name[i] = '\\';
			i+=1;
			slen+=1;
		} else {
			char ch = name[i];
			char buf[5];
			memmove(name+i+4, name+i+1, slen - i);
			sprintf(buf, "\\x%02x", ch & 0xff);
			memcpy(name+i, buf, 4);
			i+=3;
			slen+=3;
		}
	}
	return name;
}

static std::string type_name(const char *cname) {
	std::string name("type_");
	name +=(cname);
	for (auto &&c: name) {
		if (c == ' ' || c == '(' || c == ')') {
			c = '_';
		}
	}
	return name;
}

static std::string define(nvlist_t *list, std::string type_name) {
	nvpair_t *pair = NULL;
	std::stringstream def;

	def << "type " << type_name << " struct {\n";
	char field = 'A';
	while ((pair = nvlist_next_nvpair(list, pair)) != NULL) {
		auto name = sanitize(nvpair_name(pair));
		auto type = nvpair_type(pair);
		def << "\t" << field++ << " " << types[type] << " `nv:\"" << name;
		if (type == DATA_TYPE_BYTE || type == DATA_TYPE_BYTE_ARRAY) {
			def << ",byte";
		}
		def << "\"`\n";
		free(name);
	}
	def << "}\n\n";
	return def.str();
}

static void print(nvlist_t *list, char *name) {
	char *buf = NULL;
	size_t blen;
	int err;
	if ((err = nvlist_pack(list, &buf, &blen, NV_ENCODE_XDR, 0)) != 0) {
		std::cerr << "nvlist_pack error:" << err << "\n";
	}

	std::string struct_name = type_name(name);
	std::string def = define(list, struct_name);
	defs << def;
	tests << "\t{name: \"" << name << "\", ptr: func() interface{} { return &" << struct_name << "{} }, payload: []byte(\"";

	for (unsigned i = 0; i < blen; i++) {
		tests << "\\x"
			<< std::hex << std::setw(2) << std::setfill('0') << std::right
			<< (buf[i] & 0xFF);
	}
	tests << "\")},\n";
}

static char *stra(char *s, int n) {
	size_t sl = strlen(s) + 2;
	char sep[sl];
	memset(sep, '\0', sl);
	strcat(sep, s);
	strcat(sep, ";");
	sl -= 1; // remove accounting for '\0'

	size_t size = sl * n + 1; // is really 1 byte extra but makes later stuff easier
	s = (char *)calloc(1, size);
	assert(s);

	for (int i = 0; i < n; i++) {
		strcat(s, sep);
	}
	s[size - 2] = '\0'; //overwrite trailing ',' with '\0'
	return s;
}

static char *stru(unsigned long long i) {
	char *s = NULL;
	int err = asprintf(&s, "%llu", i);
	if (err == -1) {
		std::cerr << "asprintf error:" << err << "\n";
		assert(err != -1);
	}
	return s;
}

static char *stri(long long i) {
	char *s = NULL;
	int err = asprintf(&s, "%lld", i);
	if (err == -1) {
		std::cerr << "asprintf error:" << err << "\n";
		assert(err != -1);
	}
	return s;
}

static char *strf(double d) {
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
	print(l, stringify(lower) " array(" stringify(len)")"); \
	fnvlist_free(l); \
} while(0) \

int main() {
	defs <<"package nv\n"
		"\n"
		"/* !!! GENERATED FILE DO NOT EDIT !!! */\n"
		"\n"
		"import \"time\"\n\n";

	tests <<"var good = []struct {\n"
		"\tname    string\n"
		"\tptr     func() interface{}\n"
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
        fnvlist_add_boolean(l,"true");
        print(l,"boolean");
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
		arrset(array, len, B_FALSE); fnvlist_add_boolean_array(l, stra("false", len), array, len);
		arrset(array, len, B_TRUE); fnvlist_add_boolean_array(l, stra("true", len), array, len);
		print(l, "bool array");
		fnvlist_free(l);
	}

	l = fnvlist_alloc();
	//fnvlist_add_byte(l, "-128", -128);
	fnvlist_add_byte(l, "0", 0);
	fnvlist_add_byte(l, "1", 1);
	fnvlist_add_byte(l, "127", 127);
	print(l, "bytes");
	fnvlist_free(l);

	{
		l = fnvlist_alloc();
		size_t len = 5;
		unsigned char array[len];
		//arrset(array, len, -128); fnvlist_add_byte_array(l, stra("-128", len), array, len);
		arrset(array, len, 0); fnvlist_add_byte_array(l, stra("0", len), array, len);
		arrset(array, len, 1); fnvlist_add_byte_array(l, stra("1", len), array, len);
		arrset(array, len, 127); fnvlist_add_byte_array(l, stra("127", len), array, len);
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
	fnvlist_add_string(l, "01", "01");
	fnvlist_add_string(l, "012", "012");
	fnvlist_add_string(l, "0123", "0123");
	fnvlist_add_string(l, "01234", "01234");
	fnvlist_add_string(l, "012345", "012345");
	fnvlist_add_string(l, "0123456", "0123456");
	fnvlist_add_string(l, "01234567", "01234567");
	fnvlist_add_string(l, "\xff\"", "\xff\"");
	print(l, "strings");
	fnvlist_free(l);


	{
		char *array[] = {
			"0",
			"01",
			"012",
			"0123",
			"01234",
			"012345",
			"0123456",
			"01234567",
			"\xff\"",
		};
		l = fnvlist_alloc();
		fnvlist_add_string_array(l, "0;01;012;0123;01234;012345;0123456;01234567;\xff\"", array, 9);
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

	{
		l = fnvlist_alloc();
		nvlist_t *larr[] = {le, le};
		fnvlist_add_nvlist_array(l, stra("list", 2), larr, 2);
		print(l, "nvlist array");
		fnvlist_free(le);
		fnvlist_free(l);
	}

	do_double(double, DBL);

	/*
	{
		l = fnvlist_alloc();
		fnvlist_add_int8_array(l, "empty int8", {}, 0);
		fnvlist_add_int16_array(l, "empty int16", {}, 0);
		fnvlist_add_int32_array(l, "empty int32", {}, 0);
		fnvlist_add_int64_array(l, "empty int64", {}, 0);
		fnvlist_add_uint8_array(l, "empty uint8", {}, 0);
		fnvlist_add_uint16_array(l, "empty uint16", {}, 0);
		fnvlist_add_uint32_array(l, "empty uint32", {}, 0);
		fnvlist_add_uint64_array(l, "empty uint64", {}, 0);
		fnvlist_add_string_array(l, "empty string", {}, 0);
		fnvlist_add_nvlist_array(l, "empty nvlist", {}, 0);
		print(l, "empty arrays");
		fnvlist_free(l);
	}
	*/

	tests << "}\n";

	std::cout << defs.str() << tests.str() << std::flush;
	return 0;
}
