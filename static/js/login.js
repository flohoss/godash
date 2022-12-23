import { createApp, nextTick } from "https://unpkg.com/petite-vue?module";

function FieldComponent(props) {
  return {
    $template: "#field-component-template",
    field: props.field,
    get invalidMessage() {
      return props.invalidMessage();
    },
    validate(e) {
      if (e.key === "Escape") {
        this.field.value = "";
      } else {
        nextTick(() => {
          if (this.invalidMessage) props.validate();
        });
      }
    },
  };
}

function requiredFieldMessage(what) {
  return what + " is a required field";
}

createApp({
  FieldComponent,
  $delimiters: ["[[", "]]"],
  isLoading: false,
  isError: false,
  invalids: {},
  fields: {
    password: {
      label: "Password",
      type: "password",
      value: "",
      validation: { message: requiredFieldMessage("Password"), test: (value) => value },
    },
  },
  get isInvalid() {
    return !!Object.values(this.invalids).filter((key) => key).length;
  },
  validate() {
    this.invalids = {};
    Object.entries(this.fields).forEach((key) => {
      this.validateField(key[0], key[1]);
    });
  },
  validateField(fieldKey, field) {
    this.invalids[fieldKey] = false;
    if (!field.validation.test(field.value)) {
      this.invalids[fieldKey] = field.validation.message;
    }
  },
  submit() {
    this.validate();
    if (this.isInvalid) return;
    this.isLoading = true;
    const FD = new FormData();
    Object.entries(this.fields).forEach((key) => {
      FD.append(key[0], key[1].value);
    });
    fetch("/auth/login", {
      method: "post",
      body: FD,
    })
      .then((resp) => {
        if (resp.ok) {
          window.location.href = "/";
        } else this.isError = true;
      })
      .catch(() => (this.isError = true))
      .finally(() => (this.isLoading = false));
  },
}).mount("#login-form");
