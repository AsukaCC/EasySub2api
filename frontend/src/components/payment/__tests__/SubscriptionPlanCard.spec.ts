import { mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import { createPinia } from "pinia";
import { createI18n } from "vue-i18n";
import type { SubscriptionPlan } from "@/types/payment";
import SubscriptionPlanCard from "../SubscriptionPlanCard.vue";

const i18n = createI18n({
  legacy: false,
  locale: "en",
  fallbackWarn: false,
  missingWarn: false,
  messages: {
    en: {
      payment: {
        days: "days",
        weeks: "weeks",
        months: "months",
        perMonth: "month",
        models: "Models",
        planCard: {
          quota: "Quota",
          rate: "Rate",
          unlimited: "Unlimited",
        },
        subscribeNow: "Subscribe now",
      },
    },
  },
});

const mountPlanCard = (groupPlatform: string, overrides: Partial<SubscriptionPlan> = {}) =>
  mount(SubscriptionPlanCard, {
    props: {
      plan: {
        id: 1,
        group_id: 10,
        group_platform: groupPlatform,
        name: "Pro",
        price: 10,
        amount: 1000,
        features: [],
        rate_multiplier: 1,
        validity_days: 30,
        validity_unit: "day",
        is_active: true,
        ...overrides,
      },
    },
    global: { plugins: [i18n, createPinia()] },
  });

describe("SubscriptionPlanCard", () => {
  // #4607：管理端保存的单位是复数（months/weeks），此前用户侧只匹配单数
  // 'month'，「1 个月」的套餐卡片被显示成「1天」。测试环境的 vue-i18n 为
  // runtime-only 构建，t() 原样返回 key，故按 key 断言单位分支。
  it("renders plural admin-form validity units instead of mislabeled days (#4607)", () => {
    expect(mountPlanCard("openai", { validity_days: 1, validity_unit: "months" }).text()).toContain("/ payment.perMonth");
    expect(mountPlanCard("openai", { validity_days: 3, validity_unit: "months" }).text()).toContain("/ 3payment.months");
    expect(mountPlanCard("openai", { validity_days: 2, validity_unit: "weeks" }).text()).toContain("/ 2payment.weeks");
    expect(mountPlanCard("openai", { validity_days: 30, validity_unit: "day" }).text()).toContain("/ 30payment.days");
  });

  it("shows subscription prices and quotas as platform points", () => {
    const text = mountPlanCard("openai", {
      currency: "CNY",
      price_points: 18,
      original_price_points: 20,
      daily_limit_points: 50,
    }).text();

    expect(text).toContain("18.00 points");
    expect(text).toContain("20.00 points");
    expect(text).toContain("50.00 points");
    expect(text).not.toContain("¥");
    expect(text).not.toContain("$");
  });

  it.each([
    ["long Chinese", "企业全球加速专业订阅套餐（含高级模型与优先支持）"],
    ["long English", "Enterprise Global Acceleration Subscription with Priority Support"],
    ["unbroken token", "EnterpriseGlobalAccelerationSubscriptionWithPrioritySupport1234567890"],
  ])("keeps the full %s plan title accessible in a bounded two-line area", (_label, name) => {
    const wrapper = mountPlanCard("openai", { name });
    const title = wrapper.get("h3");

    expect(title.text()).toBe(name);
    expect(title.attributes("title")).toBe(name);
    expect(title.attributes("data-testid")).toBe("plan-card-title");
  });

  it("keeps title, badge, price, description, and purchase action in separate bounded regions", () => {
    const wrapper = mountPlanCard("openai", {
      name: "Enterprise Global Acceleration Subscription with Priority Support",
      price: 123.45,
      currency: "USD",
      description: "Includes advanced models and priority support.",
    });
    const title = wrapper.get("h3");
    const badge = wrapper.findAll("span").find((node) => node.text() === "OpenAI");
    const price = wrapper.findAll("span").find((node) => node.text() === "123.45 points");

    expect(title.element.parentElement?.getAttribute("data-testid")).toBe("plan-card-intro");
    expect(badge?.attributes("data-testid")).toBe("plan-card-platform");
    expect(badge?.element.parentElement?.getAttribute("data-testid")).toBe("plan-card-period-row");
    expect(badge?.element.parentElement?.textContent).toContain("/ 30payment.days");
    expect(badge?.element.parentElement?.parentElement?.getAttribute("data-testid")).toBe("plan-card-pricing");
    expect(price?.element.parentElement?.getAttribute("data-testid")).toBe("plan-card-price-row");
    expect(wrapper.get("p").text()).toBe("Includes advanced models and priority support.");
    expect(wrapper.get("button").text()).toBe("payment.subscribeNow");
  });

  it("keeps short plan titles compact and aligned", () => {
    const wrapper = mountPlanCard("openai", { name: "Pro", description: "" });
    const title = wrapper.get("h3");
    const badge = wrapper.findAll("span").find((node) => node.text() === "OpenAI");

    expect(title.text()).toBe("Pro");
    expect(title.attributes("title")).toBe("Pro");
    expect(title.attributes("data-testid")).toBe("plan-card-title");
    expect(badge?.element.parentElement?.getAttribute("data-testid")).toBe("plan-card-period-row");
    expect(badge?.element.parentElement?.textContent).toContain("/ 30payment.days");
  });
});
