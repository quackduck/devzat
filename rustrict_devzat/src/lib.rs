/// This small Rust library exposes a `censor` function that uses rustrict on a C string. 
/// The code in the helper functions is unsafe but as input is controlled and rustrict behavior
/// is known, there are no significant risks.

use rustrict::CensorStr;
use std::ffi::CStr;
use std::ffi::CString;
use std::os::raw::c_char;
use libc::memcpy;
use libc::c_void;


#[no_mangle]
pub unsafe extern "C" fn censor(pnt: *mut c_char) {
        let origin = cpnt_to_str(pnt);
        let censored: String = origin.censor();
        write_into_cpnt(pnt, &censored);
}

/// Transform a raw c_char pointer into a string.
/// Assusmes that everything went well.
fn cpnt_to_str(pnt: *const c_char) -> String {
    let c_str: &CStr = unsafe { CStr::from_ptr(pnt) };
    return c_str.to_str().unwrap().to_string()
}

/// Copies some string into a raw pointer.
/// No checks are done on size so overflows could happen.
unsafe fn write_into_cpnt(pnt: *mut c_char, s: &str) {
    let c_str = CString::new(s).unwrap();
    let c_s: *const c_char = c_str.as_ptr() as *const c_char;
    memcpy(pnt as *mut c_void, c_s as *mut c_void, s.as_bytes().len());
}

