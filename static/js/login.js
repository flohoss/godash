import { createApp, nextTick } from "https://unpkg.com/petite-vue@0.4.1/dist/petite-vue.iife.js";

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
      } else if (e.key === "Enter") {
        props.submit();
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
  invalids: {},
  fields: {
    name: {
      label: "Name",
      value: "",
      validation: { message: requiredFieldMessage("Name"), test: (value) => value },
    },
    password: {
      label: "Password",
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
    console.log("doing submit", this.fields);
  },
}).mount("#login-form");
